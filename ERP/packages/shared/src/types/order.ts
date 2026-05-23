export type OrderStatus = "open" | "paid" | "refunded" | "cancelled";
export type PaymentCode = "cash" | "kaspi_qr" | "loyalty" | "gift";

export interface OrderItemModifier {
  modifierOptionId: string;
  priceDeltaSnapshot: number;
  name?: string;
}

export interface OrderItem {
  id: string;
  productId: string;
  productName?: string;
  qty: number;
  unitPriceSnapshot: number;
  lineTotal: number;
  lineCost: number;
  modifiers: OrderItemModifier[];
}

export interface Payment {
  paymentMethodId: string;
  amount: number;
  externalRef?: string;
  code?: PaymentCode;
}

export interface Order {
  id: string;
  locationId: string;
  shiftId?: string;
  baristaId: string;
  customerId?: string;
  status: OrderStatus;
  subtotal: number;
  discountTotal: number;
  loyaltyPointsUsed: number;
  total: number;
  costTotal: number;
  receiptNo: string;
  items: OrderItem[];
  payments: Payment[];
  clientUuid: string;
  createdAt: string;
}

export interface CreateOrderRequest {
  locationId: string;
  shiftId?: string;
  customerId?: string;
  items: Array<{
    productId: string;
    qty: number;
    modifierOptionIds: string[];
  }>;
  payments: Payment[];
  loyaltyPointsUsed?: number;
  clientUuid: string;
}
