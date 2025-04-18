import { WebSocketMessage } from '../types';

class WebSocketService {
  private socket: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectTimeout = 1000; // 1 second

  connect(tableId: string, onMessage: (message: WebSocketMessage) => void) {
    if (this.socket) {
      this.disconnect();
    }

    const wsUrl = `ws://localhost:8080/ws?tableId=${tableId}`;
    this.socket = new WebSocket(wsUrl);

    this.socket.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
    };

    this.socket.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        onMessage(message);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    this.socket.onclose = () => {
      console.log('WebSocket disconnected');
      this.attemptReconnect(tableId, onMessage);
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  private attemptReconnect(tableId: string, onMessage: (message: WebSocketMessage) => void) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect(tableId, onMessage);
      }, this.reconnectTimeout * this.reconnectAttempts);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }

  disconnect() {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }
}

export const websocketService = new WebSocketService(); 