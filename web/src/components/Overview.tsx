import React from 'react';
import type { LineData, IncidentData } from '../types';

interface OverviewProps {
  lines: LineData[];
  incidents: IncidentData[];
  totalProduced: number;
  totalTarget: number;
}

export const Overview: React.FC<OverviewProps> = ({
  lines,
  incidents,
  totalProduced,
  totalTarget
}) => {
  const activeIncidents = incidents.filter(i => i.status !== 'resolved');

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '16px', marginBottom: '24px' }}>
      <div style={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', padding: '16px' }}>
        <div style={{ color: '#94a3b8', fontSize: '13px' }}>Shift Production Target</div>
        <div style={{ fontSize: '28px', fontWeight: 'bold', color: '#f8fafc', marginTop: '4px' }}>
          {totalProduced} / {totalTarget}
        </div>
        <div style={{ color: '#38bdf8', fontSize: '12px', marginTop: '4px' }}>
          {Math.round((totalProduced / totalTarget) * 100)}% of target pace
        </div>
      </div>

      {lines.map((line) => (
        <div key={line.lineId} style={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', padding: '16px' }}>
          <div style={{ color: '#94a3b8', fontSize: '13px' }}>{line.name} OEE</div>
          <div style={{ fontSize: '28px', fontWeight: 'bold', color: '#f8fafc', marginTop: '4px' }}>
            {(line.oee * 100).toFixed(1)}%
          </div>
          <div style={{ color: '#22c55e', fontSize: '12px', marginTop: '4px' }}>
            Output: {line.actualUnits} / {line.targetUnits}
          </div>
        </div>
      ))}

      <div style={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', padding: '16px' }}>
        <div style={{ color: '#94a3b8', fontSize: '13px' }}>Active Incidents</div>
        <div style={{
          fontSize: '28px',
          fontWeight: 'bold',
          color: activeIncidents.length > 0 ? '#ef4444' : '#22c55e',
          marginTop: '4px'
        }}>
          {activeIncidents.length}
        </div>
        <div style={{ color: '#94a3b8', fontSize: '12px', marginTop: '4px' }}>
          {activeIncidents.filter(i => i.status === 'open').length} Open, {activeIncidents.filter(i => i.status === 'acknowledged').length} Ack
        </div>
      </div>
    </div>
  );
};

