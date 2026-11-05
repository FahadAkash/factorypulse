using System;
using UnityEngine;

namespace FactoryPulseSim.Config
{
    [CreateAssetMenu(fileName = "NewMachineConfig", menuMenuName = "FactoryPulse/Machine Config")]
    public class MachineConfig : ScriptableObject
    {
        [Header("Identity")]
        [Tooltip("Machine ID matching contract (e.g. M-REF-01)")]
        [SerializeField] private string machineId = "M-REF-01";

        [Tooltip("1-based sequence order along the line (1 to 4)")]
        [Range(1, 4)]
        [SerializeField] private int stationOrder = 1;

        [Tooltip("Display name of this manufacturing station")]
        [SerializeField] private string stationName = "Sheet Press";

        [Header("Cycle Timing")]
        [Tooltip("Target cycle duration in milliseconds")]
        [Min(1000)]
        [SerializeField] private int idealCycleTimeMs = 50000;

        [Tooltip("Cycle duration random variance percentage (0 - 30%)")]
        [Range(0f, 30f)]
        [SerializeField] private float cycleJitterPercent = 5f;

        [Header("Quality & Defects")]
        [Tooltip("Probability of producing a rejected unit (0.0 to 1.0)")]
        [Range(0f, 1f)]
        [SerializeField] private float rejectProbability = 0.01f;

        [Tooltip("List of possible defect codes when outcome is rejected")]
        [SerializeField] private string[] defectCodes = new string[] { "DIMENSION_OUT_OF_SPEC", "SURFACE_SCRATCH" };

        [Header("Reliability (MTBF / MTTR)")]
        [Tooltip("Mean Time Between Failures in seconds (simulated)")]
        [Min(10f)]
        [SerializeField] private float mtbfSeconds = 14400f; // ~4 hours

        [Tooltip("Mean Time To Resolve / Repair in seconds (simulated)")]
        [Min(5f)]
        [SerializeField] private float mttrSeconds = 600f; // ~10 minutes

        public string MachineId => machineId;
        public int StationOrder => stationOrder;
        public string StationName => stationName;
        public int IdealCycleTimeMs => idealCycleTimeMs;
        public float CycleJitterPercent => cycleJitterPercent;
        public float RejectProbability => rejectProbability;
        public string[] DefectCodes => defectCodes;
        public float MtbfSeconds => mtbfSeconds;
        public float MttrSeconds => mttrSeconds;
    }
}

