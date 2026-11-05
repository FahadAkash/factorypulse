using System.Collections.Concurrent;
using FactoryPulseSim.Config;
using UnityEngine;

namespace FactoryPulseSim.Core
{
    /// <summary>
    /// Manages the physical topology for a single line: 4 MachineNodes connected via 3 ProductBuffers.
    /// Handles simulation lifecycle binding across all 4 stations.
    /// </summary>
    public class LineTopology : MonoBehaviour
    {
        [SerializeField] private LineConfig config;
        [SerializeField] private MachineNode[] machines = new MachineNode[4];
        [SerializeField] private ProductBuffer[] buffers = new ProductBuffer[3];

        public LineConfig Config => config;
        public MachineNode[] Machines => machines;
        public ProductBuffer[] Buffers => buffers;

        public void InitializeTopology(string factoryId, LineConfig lineCfg)
        {
            config = lineCfg;
            if (config == null)
            {
                Debug.LogError($"LineTopology on {gameObject.name} missing LineConfig!");
                return;
            }

            // Ensure 3 buffers are created or configured
            for (int i = 0; i < 3; i++)
            {
                if (buffers[i] == null)
                {
                    var bufObj = new GameObject($"Buffer_{config.LineId}_{i + 1}");
                    bufObj.transform.SetParent(transform);
                    buffers[i] = bufObj.AddComponent<ProductBuffer>();
                }
                int cap = (config.BufferCapacities != null && config.BufferCapacities.Length > i)
                    ? config.BufferCapacities[i]
                    : 10;
                buffers[i].Initialize($"B-{config.LineId}-{i + 1}", cap);
            }

            // Wire up the 4 machines to the buffers
            for (int i = 0; i < 4; i++)
            {
                if (machines[i] == null)
                {
                    var machObj = new GameObject($"Machine_{config.LineId}_{i + 1}");
                    machObj.transform.SetParent(transform);
                    machines[i] = machObj.AddComponent<MachineNode>();
                }

                ProductBuffer upstream = (i == 0) ? null : buffers[i - 1];
                ProductBuffer downstream = (i == 3) ? null : buffers[i];
                MachineConfig mCfg = (config.Machines != null && config.Machines.Length > i)
                    ? config.Machines[i]
                    : null;

                machines[i].BindConfiguration(factoryId, config.LineId, mCfg, upstream, downstream);
            }
        }

        public void InitializeSimulation(
            SimClock clock,
            int masterSeed,
            string runId,
            ConcurrentQueue<SimEvent> outbox)
        {
            if (machines == null) return;
            for (int i = 0; i < machines.Length; i++)
            {
                if (machines[i] != null)
                {
                    machines[i].InitializeSimulation(clock, masterSeed + i, runId, outbox);
                }
            }
        }
    }
}
