import { useEffect, useRef, useState, useCallback } from 'react';
import type { ConnectionStatus } from '../types';

export function useLiveSocket(onMessage?: (data: unknown) => void) {
  const [status, setStatus] = useState<ConnectionStatus>('RECONNECTING');
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const backoffRef = useRef<number>(1000);
  const lastMsgTimeRef = useRef<number>(Date.now());

  const connect = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setStatus('LIVE');
        backoffRef.current = 1000;
        lastMsgTimeRef.current = Date.now();
      };

      ws.onmessage = (event) => {
        lastMsgTimeRef.current = Date.now();
        setStatus('LIVE');
        try {
          const parsed = JSON.parse(event.data);
          onMessage?.(parsed);
        } catch {
          // ignore non-json
        }
      };

      ws.onclose = () => {
        setStatus('RECONNECTING');
        scheduleReconnect();
      };

      ws.onerror = () => {
        ws.close();
      };
    } catch {
      setStatus('RECONNECTING');
      scheduleReconnect();
    }
  }, [onMessage]);

  const scheduleReconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      window.clearTimeout(reconnectTimeoutRef.current);
    }
    const delay = Math.min(30000, backoffRef.current);
    backoffRef.current *= 1.5;
    reconnectTimeoutRef.current = window.setTimeout(() => {
      connect();
    }, delay);
  }, [connect]);

  useEffect(() => {
    connect();

    // Check for stale data if >10s without messages
    const staleInterval = window.setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        if (Date.now() - lastMsgTimeRef.current > 10000) {
          setStatus('STALE');
        } else {
          setStatus('LIVE');
        }
      }
    }, 2000);

    return () => {
      if (reconnectTimeoutRef.current) {
        window.clearTimeout(reconnectTimeoutRef.current);
      }
      window.clearInterval(staleInterval);
      wsRef.current?.close();
    };
  }, [connect]);

  return { status };
}

