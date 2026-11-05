import React from 'react';
import type { LineData, MachineState } from '../types';

interface LineViewProps {
  line: LineData;
}

const getStateColor = (state: MachineState) => {
  switch (state) {
    case 'RUNNING':
      return { bg: '#14532d', border: '#22c55e', text: '#4ade80' };
    case 'STOPPED':
      return { bg: '#7f1d1d', border: '#ef4444', text: '#fca5a5' };
    case 'STARVED':
      return { bg: '#78350f', border: '#f59e0b', text: '#fde047' };
    case 'BLOCKED':
      return { bg: '#581c87', border: '#a855f7', text: '#d8b4fe' };
    case 'IDLE':
    case 'MAINTENANCE':
    default:
      return { bg: '#334155', border: '#64748b', text: '#94a3b8' };
  }
};

export const LineView: React.FC<LineViewProps> = ({ line }) => {
  return (
    <div style={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', padding: '20px', marginBottom: '24px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: '18px', color: '#f8fafc' }}>{line.name}</h2>
          <span style={{ fontSize: '12px', color: '#94a3b8' }}>
            Product: <strong>{line.productType}</strong> | Line OEE: <strong>{(line.oee * 100).toFixed(1)}%</strong>
          </span>
        </div>
        {line.bottleneckMachineId && (
          <div style={{
            backgroundColor: '#7c2d12',
            color: '#fdba74',
            border: '1px solid #f97316',
            padding: '4px 10px',
            borderRadius: '6px',
            fontSize: '12px',
            fontWeight: '600'
          }}>
            ⚠️ Line Bottleneck: {line.bottleneckMachineId}
          </div>
        )}
      </div>

      {/* Sequential stations and buffers */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', overflowX: 'auto', paddingBottom: '8px' }}>
        {line.stations.map((st, idx) => {
          const colors = getStateColor(st.state);
          const isBottleneck = line.bottleneckMachineId === st.machineId;
          const intermediateBuffer = line.buffers[idx];

          return (
            <React.Fragment key={st.machineId}>
              {/* Station Card */}
              <div style={{
                flex: '1 0 180px',
                backgroundColor: colors.bg,
                border: `2px solid ${isBottleneck ? '#f97316' : colors.border}`,
                borderRadius: '8px',
                padding: '12px',
                boxShadow: isBottleneck ? '0 0 10px rgba(249, 115, 22, 0.4)' : 'none'
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: '11px', color: '#94a3b8' }}>#{st.stationOrder}</span>
                  <span style={{
                    fontSize: '10px',
                    fontWeight: 'bold',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    backgroundColor: colors.border,
                    color: '#0f172a'
                  }}>
                    {st.state}
                  </span>
                </div>

                <div style={{ fontWeight: 'bold', fontSize: '14px', color: '#f8fafc', marginTop: '6px' }}>
                  {st.name}
                </div>
                <div style={{ fontSize: '11px', color: '#cbd5e1' }}>{st.machineId}</div>

                {st.reasonCode && (
                  <div style={{ fontSize: '11px', color: '#fca5a5', marginTop: '6px', fontWeight: '500' }}>
                    Fault: {st.reasonCode}
                  </div>
                )}

                {st.upstreamCauseId && (
                  <div style={{ fontSize: '10px', color: '#fde047', marginTop: '4px' }}>
                    Cause: {st.upstreamCauseId}
                  </div>
                )}

                <div style={{ marginTop: '8px', paddingTop: '8px', borderTop: '1px solid rgba(255,255,255,0.1)', fontSize: '11px', color: '#cbd5e1' }}>
                  <div>Units: <strong>{st.unitsProduced}</strong></div>
                  <div>OEE: <strong>{(st.oee.overall * 100).toFixed(0)}%</strong></div>
                </div>
              </div>

              {/* Intermediate Buffer */}
              {idx < line.stations.length - 1 && intermediateBuffer && (
                <div style={{
                  flex: '0 0 80px',
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  padding: '4px'
                }}>
                  <span style={{ fontSize: '10px', color: '#94a3b8' }}>
                    {intermediateBuffer.bufferId}
                  </span>
                  <div style={{
                    width: '100%',
                    height: '10px',
                    backgroundColor: '#334155',
                    borderRadius: '5px',
                    margin: '4px 0',
                    overflow: 'hidden'
                  }}>
                    <div style={{
                      width: `${(intermediateBuffer.currentCount / intermediateBuffer.capacity) * 100}%`,
                      height: '100%',
                      backgroundColor: intermediateBuffer.currentCount >= intermediateBuffer.capacity ? '#a855f7' : '#38bdf8'
                    }} />
                  </div>
                  <span style={{ fontSize: '10px', color: '#cbd5e1' }}>
                    {intermediateBuffer.currentCount}/{intermediateBuffer.capacity}
                  </span>
                </div>
              )}
            </React.Fragment>
          );
        })}
      </div>
    </div>
  );
};

