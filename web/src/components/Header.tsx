import React from 'react';
import type { ConnectionStatus } from '../types';

interface HeaderProps {
  status: ConnectionStatus;
  activeShift: string;
}

export const Header: React.FC<HeaderProps> = ({ status, activeShift }) => {
  const getBadgeStyle = () => {
    switch (status) {
      case 'LIVE':
        return { backgroundColor: '#14532d', color: '#4ade80', border: '1px solid #22c55e' };
      case 'RECONNECTING':
        return { backgroundColor: '#78350f', color: '#fde047', border: '1px solid #f59e0b' };
      case 'STALE':
        return { backgroundColor: '#7f1d1d', color: '#fca5a5', border: '1px solid #ef4444' };
    }
  };

  return (
    <header style={{
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      padding: '16px 24px',
      backgroundColor: '#1e293b',
      borderBottom: '1px solid #334155'
    }}>
      <div>
        <h1 style={{ margin: 0, fontSize: '20px', fontWeight: 'bold', color: '#f8fafc' }}>
          FactoryPulse
        </h1>
        <span style={{ fontSize: '13px', color: '#94a3b8' }}>
          Nova Appliances Ltd — Central Plant (factory-01)
        </span>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
        <div style={{ fontSize: '13px', color: '#cbd5e1' }}>
          <strong>Shift:</strong> {activeShift}
        </div>
        <span style={{
          padding: '4px 10px',
          borderRadius: '9999px',
          fontSize: '12px',
          fontWeight: '600',
          textTransform: 'uppercase',
          ...getBadgeStyle()
        }}>
          ● {status === 'STALE' ? 'STALE (>10s)' : status}
        </span>
      </div>
    </header>
  );
};

