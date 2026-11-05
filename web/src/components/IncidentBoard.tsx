import React from 'react';
import type { IncidentData } from '../types';

interface IncidentBoardProps {
  incidents: IncidentData[];
  onAcknowledge: (id: string) => void;
  onResolve: (id: string) => void;
}

export const IncidentBoard: React.FC<IncidentBoardProps> = ({
  incidents,
  onAcknowledge,
  onResolve
}) => {
  return (
    <div style={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px', padding: '20px', marginBottom: '24px' }}>
      <h2 style={{ margin: '0 0 16px 0', fontSize: '18px', color: '#f8fafc' }}>
        Downtime Incident Response Board
      </h2>

      {incidents.length === 0 ? (
        <div style={{ color: '#22c55e', padding: '16px 0', fontSize: '14px' }}>
          ✓ No active downtime incidents. All machines operational.
        </div>
      ) : (
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '13px', color: '#cbd5e1' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid #334155', textAlign: 'left', color: '#94a3b8' }}>
                <th style={{ padding: '8px' }}>Machine</th>
                <th style={{ padding: '8px' }}>Status</th>
                <th style={{ padding: '8px' }}>Root Cause</th>
                <th style={{ padding: '8px' }}>Opened At</th>
                <th style={{ padding: '8px' }}>Assignee</th>
                <th style={{ padding: '8px', textAlign: 'right' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {incidents.map((inc) => (
                <tr key={inc.id} style={{ borderBottom: '1px solid #334155' }}>
                  <td style={{ padding: '10px 8px', fontWeight: 'bold', color: '#f8fafc' }}>
                    {inc.machineId}
                  </td>
                  <td style={{ padding: '10px 8px' }}>
                    <span style={{
                      padding: '2px 8px',
                      borderRadius: '4px',
                      fontSize: '11px',
                      fontWeight: 'bold',
                      textTransform: 'uppercase',
                      backgroundColor: inc.status === 'open' ? '#7f1d1d' : inc.status === 'acknowledged' ? '#78350f' : '#14532d',
                      color: inc.status === 'open' ? '#fca5a5' : inc.status === 'acknowledged' ? '#fde047' : '#4ade80'
                    }}>
                      {inc.status}
                    </span>
                  </td>
                  <td style={{ padding: '10px 8px' }}>{inc.rootCause || 'Under Diagnosis'}</td>
                  <td style={{ padding: '10px 8px' }}>{new Date(inc.openedAt).toLocaleTimeString()}</td>
                  <td style={{ padding: '10px 8px' }}>{inc.assignee || 'Unassigned'}</td>
                  <td style={{ padding: '10px 8px', textAlign: 'right' }}>
                    {inc.status === 'open' && (
                      <button
                        onClick={() => onAcknowledge(inc.id)}
                        style={{
                          backgroundColor: '#f59e0b',
                          color: '#0f172a',
                          border: 'none',
                          padding: '4px 10px',
                          borderRadius: '4px',
                          fontWeight: '600',
                          cursor: 'pointer',
                          marginRight: '6px'
                        }}
                      >
                        Acknowledge
                      </button>
                    )}
                    {inc.status !== 'resolved' && (
                      <button
                        onClick={() => onResolve(inc.id)}
                        style={{
                          backgroundColor: '#22c55e',
                          color: '#0f172a',
                          border: 'none',
                          padding: '4px 10px',
                          borderRadius: '4px',
                          fontWeight: '600',
                          cursor: 'pointer'
                        }}
                      >
                        Resolve
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

