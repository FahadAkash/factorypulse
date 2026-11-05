using System;
using FactoryPulseSim.Config;
using UnityEngine;

namespace FactoryPulseSim.Core
{
    /// <summary>
    /// Root manager for the simulated factory: 3 lines x 4 connected machines + buffers.
    /// In M3, this constructs and verifies the full plant topology from FactoryConfig.
    /// </summary>
    public class FactoryTopology : MonoBehaviour
    {
        [Header("Master Factory Configuration")]
        [SerializeField] private FactoryConfig config;

        [Header("Simulation Clock")]
        [SerializeField] private SimClock clock;

        [Header("Line Topologies (3 Lines)")]
        [SerializeField] private LineTopology[] lines = new LineTopology[3];

        public FactoryConfig Config => config;
        public SimClock Clock => clock;
        public LineTopology[] Lines => lines;

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

        public void InitializeFactory(FactoryConfig factoryCfg)
        {
            config = factoryCfg;
            if (config == null)
            {
                Debug.LogError("FactoryTopology missing FactoryConfig!");
                return;
            }

            clock.Initialize(DateTime.UtcNow, 1.0f);

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
            }
        }
    }
}
