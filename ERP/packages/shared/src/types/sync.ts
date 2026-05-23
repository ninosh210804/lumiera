export type SyncStatus = "online" | "offline" | "syncing" | "needs_review";

export type SyncEventType =
  | "ORDER_CREATED"
  | "ORDER_REFUNDED"
  | "ORDER_CANCELLED"
  | "STOCK_RECEIVED"
  | "STOCK_ADJUSTED"
  | "STOCK_WRITTEN_OFF"
  | "PRICE_CHANGED"
  | "MENU_PUBLISHED"
  | "SHIFT_OPENED"
  | "SHIFT_CLOSED"
  | "CASH_DRAWER_OP"
  | "LOYALTY_EARNED"
  | "LOYALTY_SPENT";

export interface SyncEvent {
  clientUuid: string;
  sequence: number;
  eventType: SyncEventType;
  payload: Record<string, unknown>;
  deviceTs: string;
}

export interface SyncPushResponse {
  accepted: string[];
  conflicts: Array<{
    clientUuid: string;
    kind: string;
  }>;
}

export interface SyncPullResponse {
  events: SyncEvent[];
  nextCursor: string;
  hasMore: boolean;
}
