using System;
using UnityEngine;

namespace FactoryPulseSim.Config
{
    [CreateAssetMenu(fileName = "NewLineConfig", menuMenuName = "FactoryPulse/Line Config")]
    public class LineConfig : ScriptableObject
    {
        [Header("Identity")]
        [Tooltip("Line ID matching contract (e.g. line-refrig-01)")]
        [SerializeField] private string lineId = "line-refrig-01";

        [Tooltip("Descriptive line name")]
        [SerializeField] private string lineName = "Refrigerator Assembly Line";

        [Tooltip("Product type identifier (e.g. refrigerator-standard)")]
        [SerializeField] private string productType = "refrigerator-standard";

        [Header("Targets")]
        [Tooltip("Target units per 8-hour shift")]
        [Min(1)]
        [SerializeField] private int targetShiftUnits = 500;

        [Header("Topology (4 Connected Machines)")]
        [Tooltip("Ordered station configurations (exactly 4 stations)")]
        [SerializeField] private MachineConfig[] machines = new MachineConfig[4];

        [Header("Intermediate Buffers (3 Buffers)")]
        [Tooltip("Capacities for buffers between stations 1-2, 2-3, and 3-4")]
        [SerializeField] private int[] bufferCapacities = new int[] { 10, 8, 12 };

        public string LineId => lineId;
        public string LineName => lineName;
        public string ProductType => productType;
        public int TargetShiftUnits => targetShiftUnits;
        public MachineConfig[] Machines => machines;
        public int[] BufferCapacities => bufferCapacities;

        public void ValidateTopology()
        {
            if (machines == null || machines.Length != 4)
            {
                Debug.LogWarning($"LineConfig {lineId} must have exactly 4 MachineConfigs.");
            }
            if (bufferCapacities == null || bufferCapacities.Length != 3)
            {
                Debug.LogWarning($"LineConfig {lineId} must have exactly 3 buffer capacity definitions.");
            }
        }
    }
}

