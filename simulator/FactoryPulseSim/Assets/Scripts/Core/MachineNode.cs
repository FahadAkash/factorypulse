using System;
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
    /// Represents an individual manufacturing station connected to upstream and downstream buffers.
    /// In M3, this establishes the topological connections and state properties.
    /// Full cycle execution, fault injection, and event emission are layered in M4/M5.
    /// </summary>
    public class MachineNode : MonoBehaviour
    {
        [Header("Configuration")]
        [SerializeField] private MachineConfig config;
        [SerializeField] private string factoryId = "factory-01";
        [SerializeField] private string lineId = "line-refrig-01";

        [Header("Buffer Connections")]
        [Tooltip("Upstream buffer (null if first station in line)")]
        [SerializeField] private ProductBuffer upstreamBuffer;

        [Tooltip("Downstream buffer (null if last station in line)")]
        [SerializeField] private ProductBuffer downstreamBuffer;

        [Header("Runtime State")]
        [SerializeField] private MachineOperationalState currentState = MachineOperationalState.IDLE;
        [SerializeField] private string currentReasonCode = null;
        [SerializeField] private string upstreamCauseMachineId = null;
        [SerializeField] private long sequenceNumber = 0;

        public event Action<MachineNode, MachineOperationalState> OnStateChanged;

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

        public bool IsFirstStation => upstreamBuffer == null;
        public bool IsLastStation => downstreamBuffer == null;

        public void BindConfiguration(string fId, string lId, MachineConfig machineCfg, ProductBuffer upstream, ProductBuffer downstream)
        {
            factoryId = fId;
            lineId = lId;
            config = machineCfg;
            upstreamBuffer = upstream;
            downstreamBuffer = downstream;
            currentState = MachineOperationalState.IDLE;
            sequenceNumber = 0;
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

            OnStateChanged?.Invoke(this, currentState);
        }

        public void ResetStation()
        {
            currentState = MachineOperationalState.IDLE;
            currentReasonCode = null;
            upstreamCauseMachineId = null;
            sequenceNumber = 0;
        }
    }
}
