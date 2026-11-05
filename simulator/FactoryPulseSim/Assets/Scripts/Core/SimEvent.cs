using System;
using System.Text;
using FactoryPulseSim.Utils;

namespace FactoryPulseSim.Core
{
    /// <summary>
    /// Thread-safe, non-allocating simulator event envelope matching contracts/event-schema.json.
    /// </summary>
    public readonly struct SimEvent
    {
        public readonly int SchemaVersion;
        public readonly string EventId;
        public readonly string EventType;
        public readonly string FactoryId;
        public readonly string LineId;
        public readonly string MachineId;
        public readonly string RunId;
        public readonly long Seq;
        public readonly DateTime OccurredAt;
        public readonly string PayloadJson;

        public SimEvent(
            string eventType,
            string factoryId,
            string lineId,
            string machineId,
            string runId,
            long seq,
            DateTime occurredAt,
            string payloadJson)
        {
            SchemaVersion = 1;
            EventType = eventType;
            FactoryId = factoryId;
            LineId = lineId;
            MachineId = machineId;
            RunId = runId;
            Seq = seq;
            OccurredAt = occurredAt;
            PayloadJson = payloadJson;

            // Deterministic UUIDv5 from 'run_id:machine_id:seq'
            string nameKey = $"{runId}:{machineId}:{seq}";
            EventId = UuidV5.Create(nameKey).ToString();
        }

        public string ToJsonEnvelope()
        {
            var sb = new StringBuilder(256 + PayloadJson.Length);
            sb.Append("{");
            sb.Append("\"schema_version\":1,");
            sb.Append($"\"event_id\":\"{EventId}\",");
            sb.Append($"\"event_type\":\"{EventType}\",");
            sb.Append($"\"factory_id\":\"{FactoryId}\",");
            sb.Append($"\"line_id\":\"{LineId}\",");
            sb.Append($"\"machine_id\":\"{MachineId}\",");
            sb.Append($"\"run_id\":\"{RunId}\",");
            sb.Append($"\"seq\":{Seq},");
            sb.Append($"\"occurred_at\":\"{OccurredAt:O}\",");
            sb.Append($"\"payload\":{PayloadJson}");
            sb.Append("}");
            return sb.ToString();
        }
    }
}
