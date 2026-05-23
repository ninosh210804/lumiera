export type RoleCode = "admin" | "manager" | "barista";

export interface User {
  id: string;
  fullName: string;
  email?: string;
  roleCode: RoleCode;
  defaultLocationId: string;
  isActive: boolean;
}

export interface Location {
  id: string;
  name: string;
  address: string;
  city: string;
  timezone: string;
  phone?: string;
  isActive: boolean;
}

export interface AuthTokenPayload {
  userId: string;
  locationId: string;
  deviceId?: string;
  role: RoleCode;
  exp: number;
}
