using System;
using UnityEngine;

namespace FactoryPulseSim.Config
{
    [CreateAssetMenu(fileName = "NewFactoryConfig", menuMenuName = "FactoryPulse/Factory Config")]
    public class FactoryConfig : ScriptableObject
    {
        [Header("Factory Identity")]
        [Tooltip("Factory ID matching contract (e.g. factory-01)")]
        [SerializeField] private string factoryId = "factory-01";

        [Tooltip("Plant location or name")]
        [SerializeField] private string factoryName = "Nova Central Plant";

        [Header("Production Lines (3 Lines)")]
        [Tooltip("Configurations for the 3 production lines")]
        [SerializeField] private LineConfig[] lines = new LineConfig[3];

        [Header("Shift Parameters")]
        [Tooltip("Total shift duration in minutes (standard: 480m / 8h)")]
        [Min(60)]
        [SerializeField] private int shiftDurationMinutes = 480;

        [Tooltip("Net planned operating time in minutes (excluding planned breaks: 450m)")]
        [Min(30)]
        [SerializeField] private int plannedProductionMinutes = 450;

        [Header("MQTT Broker Default Settings")]
        [SerializeField] private string defaultBrokerHost = "localhost";
        [SerializeField] private int defaultBrokerPort = 1883;

        [Header("Simulation Seed & Determinism")]
        [Tooltip("Master random seed for reproducible scenario runs")]
        [SerializeField] private int runSeed = 42;

        public string FactoryId => factoryId;
        public string FactoryName => factoryName;
        public LineConfig[] Lines => lines;
        public int ShiftDurationMinutes => shiftDurationMinutes;
        public int PlannedProductionMinutes => plannedProductionMinutes;
        public string DefaultBrokerHost => defaultBrokerHost;
        public int DefaultBrokerPort => defaultBrokerPort;
        public int RunSeed => runSeed;
    }
}

