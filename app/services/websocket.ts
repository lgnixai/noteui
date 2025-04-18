import { WS_BASE_URL } from '../config';
import { RecordUpdateMessage } from '../types/record';

export type WebSocketMessageHandler = (message: RecordUpdateMessage) => void;

export class WebSocketService {
  private static instance: WebSocketService;
  private ws: WebSocket | null = null;
  private messageHandlers: Set<WebSocketMessageHandler> = new Set();
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 5;
  private reconnectTimeout: number = 1000; // Start with 1 second

  private constructor() {
    // Private constructor for singleton pattern
  }

  public static getInstance(): WebSocketService {
    if (!WebSocketService.instance) {
      WebSocketService.instance = new WebSocketService();
    }
    return WebSocketService.instance;
  }

  public connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return; // Already connected
    }

    this.ws = new WebSocket(WS_BASE_URL);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.reconnectTimeout = 1000;
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data) as RecordUpdateMessage;
        this.messageHandlers.forEach(handler => handler(message));
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.ws = null;

      // Attempt to reconnect if we haven't exceeded max attempts
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        setTimeout(() => {
          this.reconnectAttempts++;
          this.reconnectTimeout *= 2; // Exponential backoff
          this.connect();
        }, this.reconnectTimeout);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  }

  public disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.messageHandlers.clear();
  }

  public subscribe(handler: WebSocketMessageHandler): () => void {
    this.messageHandlers.add(handler);

    // Return unsubscribe function
    return () => {
      this.messageHandlers.delete(handler);
    };
  }

  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export const webSocketService = WebSocketService.getInstance(); 