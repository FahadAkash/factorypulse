using System;
using System.Collections.Concurrent;
using FactoryPulseSim.Config;
using UnityEngine;

namespace FactoryPulseSim.Core
{
    /// <summary>
    /// Root manager for the simulated factory: 3 lines x 4 connected machines + buffers.
    /// Manages master clock, outbox queue for MQTT emission, and run lifecycle.
    /// </summary>
    public class FactoryTopology : MonoBehaviour
    {
        [Header("Master Factory Configuration")]
        [SerializeField] private FactoryConfig config;

        [Header("Simulation Clock")]
        [SerializeField] private SimClock clock;

        [Header("Line Topologies (3 Lines)")]
        [SerializeField] private LineTopology[] lines = new LineTopology[3];

        [Header("Run Information")]
        [SerializeField] private string currentRunId = "run-001";

        private readonly ConcurrentQueue<SimEvent> _eventOutbox = new ConcurrentQueue<SimEvent>();

        public FactoryConfig Config => config;
        public SimClock Clock => clock;
        public LineTopology[] Lines => lines;
        public string CurrentRunId => currentRunId;
        public ConcurrentQueue<SimEvent> EventOutbox => _eventOutbox;

        private void Awake()
        {
            if (clock == null)
            {
                clock = GetComponentInChildren<SimClock>();
                if (clock == null)
                {
                    var clockObj = new GameObject("SimClock");
                    clockObj.transform.SetParent(transform);
                    clock = clockObj.AddComponent<SimClock>();
                }
            }

            if (config != null)
            {
                InitializeFactory(config);
            }
        }

        public void InitializeFactory(FactoryConfig factoryCfg, string runId = null)
        {
            config = factoryCfg;
            if (config == null)
            {
                Debug.LogError("FactoryTopology missing FactoryConfig!");
                return;
            }

            currentRunId = !string.IsNullOrEmpty(runId) ? runId : $"run-{DateTime.UtcNow:yyyyMMdd-HHmmss}";
            clock.Initialize(DateTime.UtcNow, 1.0f);

            int masterSeed = config.RunSeed;

            for (int i = 0; i < 3; i++)
            {
                if (lines[i] == null)
                {
                    string lineName = (config.Lines != null && config.Lines.Length > i && config.Lines[i] != null)
                        ? config.Lines[i].LineId
                        : $"Line_{i + 1}";
                    var lineObj = new GameObject($"LineTopology_{lineName}");
                    lineObj.transform.SetParent(transform);
                    lines[i] = lineObj.AddComponent<LineTopology>();
                }

                LineConfig lCfg = (config.Lines != null && config.Lines.Length > i) ? config.Lines[i] : null;
                lines[i].InitializeTopology(config.FactoryId, lCfg);
                lines[i].InitializeSimulation(clock, masterSeed + (i * 10), currentRunId, _eventOutbox);
            }
        }

        public void InjectFault(string machineId, string reason, float durationSeconds)
        {
            var machine = FindMachine(machineId);
            if (machine != null)
            {
                machine.InjectManualFault(reason, durationSeconds);
            }
            else
            {
                Debug.LogWarning($"Machine {machineId} not found for fault injection.");
            }
        }

        public void ClearFault(string machineId)
        {
            var machine = FindMachine(machineId);
            if (machine != null)
            {
                machine.ClearFault();
            }
        }

        public MachineNode FindMachine(string machineId)
        {
            if (lines == null) return null;
            foreach (var line in lines)
            {
                if (line == null || line.Machines == null) continue;
                foreach (var machine in line.Machines)
                {
                    if (machine != null && machine.MachineId == machineId)
                    {
                        return machine;
                    }
                }
            }
            return null;
        }
    }
}
