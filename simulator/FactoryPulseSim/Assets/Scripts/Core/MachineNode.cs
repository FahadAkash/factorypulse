using System;
using System.Collections.Concurrent;
using System.Text;
using FactoryPulseSim.Config;
using UnityEngine;

namespace FactoryPulseSim.Core
{
    public enum MachineOperationalState
    {
        IDLE,
        RUNNING,
        STOPPED,
        STARVED,
        BLOCKED,
        MAINTENANCE
    }

    /// <summary>
    /// Executes deterministic station processing, starvation/blocking propagation,
    /// unit production, and event generation matching contracts/event-schema.json.
    /// </summary>
    public class MachineNode : MonoBehaviour
    {
        [Header("Configuration")]
        [SerializeField] private MachineConfig config;
        [SerializeField] private string factoryId = "factory-01";
        [SerializeField] private string lineId = "line-refrig-01";
        [SerializeField] private string runId = "run-001";

        [Header("Buffer Connections")]
        [SerializeField] private ProductBuffer upstreamBuffer;
        [SerializeField] private ProductBuffer downstreamBuffer;

        [Header("Runtime State")]
        [SerializeField] private MachineOperationalState currentState = MachineOperationalState.IDLE;
        [SerializeField] private string currentReasonCode = null;
        [SerializeField] private string upstreamCauseMachineId = null;
        [SerializeField] private long sequenceNumber = 0;

        [Header("Current Cycle Progress")]
        [SerializeField] private bool hasWorkPiece = false;
        [SerializeField] private float currentCycleElapsedMs = 0f;
        [SerializeField] private float currentTargetCycleMs = 50000f;
        [SerializeField] private bool currentUnitIsGood = true;
        [SerializeField] private string currentDefectCode = null;

        private SimClock _clock;
        private FailureInjector _failureInjector;
        private System.Random _cycleRng;
        private ConcurrentQueue<SimEvent> _eventOutbox;

        public event Action<MachineNode, MachineOperationalState> OnStateChanged;
        public event Action<MachineNode, SimEvent> OnEventEmitted;

        public MachineConfig Config => config;
        public string FactoryId => factoryId;
        public string LineId => lineId;
        public string MachineId => config != null ? config.MachineId : name;
        public int StationOrder => config != null ? config.StationOrder : 1;
        public ProductBuffer UpstreamBuffer => upstreamBuffer;
        public ProductBuffer DownstreamBuffer => downstreamBuffer;
        public MachineOperationalState CurrentState => currentState;
        public string CurrentReasonCode => currentReasonCode;
        public string UpstreamCauseMachineId => upstreamCauseMachineId;
        public long SequenceNumber => sequenceNumber;
        public bool HasWorkPiece => hasWorkPiece;

        public void BindConfiguration(
            string fId,
            string lId,
            MachineConfig machineCfg,
            ProductBuffer upstream,
            ProductBuffer downstream)
        {
            factoryId = fId;
            lineId = lId;
            config = machineCfg;
            upstreamBuffer = upstream;
            downstreamBuffer = downstream;
            currentState = MachineOperationalState.IDLE;
            sequenceNumber = 0;
            hasWorkPiece = false;
        }

        public void InitializeSimulation(
            SimClock simClock,
            int masterSeed,
            string currentRunId,
            ConcurrentQueue<SimEvent> outbox)
        {
            _clock = simClock;
            runId = currentRunId;
            _eventOutbox = outbox;

            int machineSeed = unchecked(masterSeed * 397 ^ MachineId.GetHashCode());
            _cycleRng = new System.Random(machineSeed);

            float mtbf = config != null ? config.MtbfSeconds : 14400f;
            float mttr = config != null ? config.MttrSeconds : 600f;
            string[] reasonCodes = config != null ? config.DefectCodes : null;

            _failureInjector = new FailureInjector(masterSeed, MachineId, mtbf, mttr, reasonCodes);
            ResetStation();
        }

        public void SetState(MachineOperationalState newState, string reason = null, string upstreamCause = null)
        {
            if (currentState == newState && currentReasonCode == reason && upstreamCauseMachineId == upstreamCause)
            {
                return;
            }

            currentState = newState;
            currentReasonCode = reason;
            upstreamCauseMachineId = upstreamCause;
            sequenceNumber++;

            // Emit machine.state_changed event
            DateTime eventTime = _clock != null ? _clock.CurrentSimTime : DateTime.UtcNow;
            string payloadJson = BuildStatePayload(newState, reason, upstreamCause);
            var simEv = new SimEvent(
                "machine.state_changed",
                factoryId,
                lineId,
                MachineId,
                runId,
                sequenceNumber,
                eventTime,
                payloadJson);

            _eventOutbox?.Enqueue(simEv);
            OnStateChanged?.Invoke(this, currentState);
            OnEventEmitted?.Invoke(this, simEv);
        }

        public void InjectManualFault(string reason, float durationSeconds)
        {
            if (_failureInjector != null && _clock != null)
            {
                _failureInjector.InjectManualFault(_clock.ElapsedSimSeconds, reason, durationSeconds);
            }
        }

        public void ClearFault()
        {
            if (_failureInjector != null && _clock != null)
            {
                _failureInjector.ClearFault(_clock.ElapsedSimSeconds);
            }
        }

        public void ResetStation()
        {
            currentState = MachineOperationalState.IDLE;
            currentReasonCode = null;
            upstreamCauseMachineId = null;
            sequenceNumber = 0;
            hasWorkPiece = false;
            currentCycleElapsedMs = 0f;
        }

        private void FixedUpdate()
        {
            if (_clock == null || !_clock.IsRunning || config == null) return;

            double simSeconds = _clock.ElapsedSimSeconds;
            float stepDeltaMs = _clock.StepDeltaSeconds * _clock.TimeScale * 1000f;

            // 1. Check Failure Injector for STOPPED state
            if (_failureInjector != null && _failureInjector.UpdateFailureStatus(simSeconds, out string faultReason))
            {
                SetState(MachineOperationalState.STOPPED, faultReason, null);
                return;
            }

            // 2. If finished unit is waiting to be pushed downstream (BLOCKED condition)
            if (hasWorkPiece && currentCycleElapsedMs >= currentTargetCycleMs)
            {
                if (downstreamBuffer != null && downstreamBuffer.IsFull)
                {
                    string causeId = FindDownstreamRootCause();
                    SetState(MachineOperationalState.BLOCKED, null, causeId);
                    return;
                }

                // Buffer has room or last station (finished goods pallet)
                if (downstreamBuffer == null || downstreamBuffer.TryPush())
                {
                    // Emit production.unit event
                    EmitProductionEvent();
                    hasWorkPiece = false;
                    currentCycleElapsedMs = 0f;
                }
            }

            // 3. If idle and waiting for incoming material (STARVED condition)
            if (!hasWorkPiece)
            {
                if (upstreamBuffer != null && upstreamBuffer.IsEmpty)
                {
                    string causeId = FindUpstreamRootCause();
                    SetState(MachineOperationalState.STARVED, null, causeId);
                    return;
                }

                // First station has infinite incoming material, or pop unit from upstream buffer
                if (upstreamBuffer == null || upstreamBuffer.TryPop())
                {
                    StartNewCycle();
                }
                else
                {
                    return;
                }
            }

            // 4. Actively processing work piece (RUNNING)
            if (hasWorkPiece)
            {
                SetState(MachineOperationalState.RUNNING, null, null);
                currentCycleElapsedMs += stepDeltaMs;
            }
        }

        private void StartNewCycle()
        {
            hasWorkPiece = true;
            currentCycleElapsedMs = 0f;

            // Compute ideal cycle time with configured jitter
            float baseMs = config.IdealCycleTimeMs;
            float jitterPct = config.CycleJitterPercent / 100f;
            float jitterFactor = 1.0f + (float)((_cycleRng.NextDouble() * 2.0 - 1.0) * jitterPct);
            currentTargetCycleMs = Mathf.Max(100f, baseMs * jitterFactor);

            // Determine unit quality
            double qualityRoll = _cycleRng.NextDouble();
            if (qualityRoll < config.RejectProbability)
            {
                currentUnitIsGood = false;
                if (config.DefectCodes != null && config.DefectCodes.Length > 0)
                {
                    currentDefectCode = config.DefectCodes[_cycleRng.Next(config.DefectCodes.Length)];
                }
                else
                {
                    currentDefectCode = "GENERAL_DEFECT";
                }
            }
            else
            {
                currentUnitIsGood = true;
                currentDefectCode = null;
            }
        }

        private void EmitProductionEvent()
        {
            sequenceNumber++;
            DateTime eventTime = _clock != null ? _clock.CurrentSimTime : DateTime.UtcNow;

            string outcome = currentUnitIsGood ? "good" : "rejected";
            int cycleTime = Mathf.RoundToInt(currentTargetCycleMs);
            string prodType = config.StationName;

            var sb = new StringBuilder(128);
            sb.Append("{");
            sb.Append($"\"outcome\":\"{outcome}\",");
            sb.Append($"\"cycle_time_ms\":{cycleTime},");
            sb.Append($"\"product_type\":\"{prodType}\"");
            if (!currentUnitIsGood && !string.IsNullOrEmpty(currentDefectCode))
            {
                sb.Append($",\"defect_code\":\"{currentDefectCode}\"");
            }
            sb.Append("}");

            var simEv = new SimEvent(
                "production.unit",
                factoryId,
                lineId,
                MachineId,
                runId,
                sequenceNumber,
                eventTime,
                sb.ToString());

            _eventOutbox?.Enqueue(simEv);
            OnEventEmitted?.Invoke(this, simEv);
        }

        private string BuildStatePayload(MachineOperationalState state, string reason, string upstreamCause)
        {
            var sb = new StringBuilder(128);
            sb.Append("{");
            sb.Append($"\"state\":\"{state}\"");
            if (state == MachineOperationalState.STOPPED && !string.IsNullOrEmpty(reason))
            {
                sb.Append($",\"reason_code\":\"{reason}\"");
            }
            if (!string.IsNullOrEmpty(upstreamCause))
            {
                sb.Append($",\"upstream_cause_machine_id\":\"{upstreamCause}\"");
            }
            sb.Append("}");
            return sb.ToString();
        }

        private string FindUpstreamRootCause()
        {
            if (transform.parent == null) return null;
            var lineTopo = transform.parent.GetComponent<LineTopology>();
            if (lineTopo == null) return null;

            int myIdx = StationOrder - 1;
            for (int i = myIdx - 1; i >= 0; i--)
            {
                if (lineTopo.Machines != null && lineTopo.Machines.Length > i && lineTopo.Machines[i] != null)
                {
                    if (lineTopo.Machines[i].CurrentState == MachineOperationalState.STOPPED)
                    {
                        return lineTopo.Machines[i].MachineId;
                    }
                }
            }
            return null;
        }

        private string FindDownstreamRootCause()
        {
            if (transform.parent == null) return null;
            var lineTopo = transform.parent.GetComponent<LineTopology>();
            if (lineTopo == null) return null;

            int myIdx = StationOrder - 1;
            for (int i = myIdx + 1; i < lineTopo.Machines.Length; i++)
            {
                if (lineTopo.Machines != null && lineTopo.Machines.Length > i && lineTopo.Machines[i] != null)
                {
                    if (lineTopo.Machines[i].CurrentState == MachineOperationalState.STOPPED)
                    {
                        return lineTopo.Machines[i].MachineId;
                    }
                }
            }
            return null;
        }
    }
}
