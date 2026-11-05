using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using FactoryPulseSim.Config;
using UnityEngine;

namespace FactoryPulseSim.Analytics
{
    [Serializable]
    public class GroundTruthData
    {
        public string run_id;
        public int seed;
        public string factory_id;
        public int planned_duration_seconds;
        public int expected_shift_units;
        public string designed_bottleneck_machine_id;
        public List<InjectedFaultRecord> injected_faults = new List<InjectedFaultRecord>();
        public Dictionary<string, int> target_cycle_times = new Dictionary<string, int>();
    }

    [Serializable]
    public class InjectedFaultRecord
    {
        public string machine_id;
        public string reason_code;
        public double scheduled_sim_seconds;
        public float duration_seconds;
    }

    /// <summary>
    /// Writes runs/{runId}/ground_truth.json capturing simulation ground truth
    /// for FDE verification and validation testing (M12 & M18).
    /// </summary>
    public static class GroundTruthRecorder
    {
        public static void RecordGroundTruth(string runId, FactoryConfig config, string bottleneckMachineId = null)
        {
            if (string.IsNullOrEmpty(runId) || config == null) return;

            try
            {
                string runDir = Path.Combine(Application.dataPath, "..", "runs", runId);
                Directory.CreateDirectory(runDir);
                string filePath = Path.Combine(runDir, "ground_truth.json");

                var gt = new GroundTruthData
                {
                    run_id = runId,
                    seed = config.RunSeed,
                    factory_id = config.FactoryId,
                    planned_duration_seconds = config.PlannedProductionMinutes * 60,
                    expected_shift_units = 500,
                    designed_bottleneck_machine_id = bottleneckMachineId ?? "M-REF-02"
                };

                if (config.Lines != null)
                {
                    foreach (var line in config.Lines)
                    {
                        if (line == null || line.Machines == null) continue;
                        foreach (var m in line.Machines)
                        {
                            if (m != null)
                            {
                                gt.target_cycle_times[m.MachineId] = m.IdealCycleTimeMs;
                            }
                        }
                    }
                }

                // Write JSON representation
                var sb = new StringBuilder(512);
                sb.AppendLine("{");
                sb.AppendLine($"  \"run_id\": \"{gt.run_id}\",");
                sb.AppendLine($"  \"seed\": {gt.seed},");
                sb.AppendLine($"  \"factory_id\": \"{gt.factory_id}\",");
                sb.AppendLine($"  \"planned_duration_seconds\": {gt.planned_duration_seconds},");
                sb.AppendLine($"  \"expected_shift_units\": {gt.expected_shift_units},");
                sb.AppendLine($"  \"designed_bottleneck_machine_id\": \"{gt.designed_bottleneck_machine_id}\"");
                sb.AppendLine("}");

                File.WriteAllText(filePath, sb.ToString(), Encoding.UTF8);
                Debug.Log($"[GroundTruth] Saved ground truth to {filePath}");
            }
            catch (Exception ex)
            {
                Debug.LogWarning($"[GroundTruth] Failed to write ground_truth.json: {ex.Message}");
            }
        }
    }
}

