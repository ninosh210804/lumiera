import { format, formatDistanceToNow, parseISO } from "date-fns";
import { ru } from "date-fns/locale";

export function formatDate(iso: string, fmt = "dd.MM.yyyy"): string {
  return format(parseISO(iso), fmt, { locale: ru });
}

export function formatDateTime(iso: string): string {
  return format(parseISO(iso), "dd.MM.yyyy HH:mm", { locale: ru });
}

export function formatRelative(iso: string): string {
  return formatDistanceToNow(parseISO(iso), { addSuffix: true, locale: ru });
}

export function toISOString(date: Date = new Date()): string {
  return date.toISOString();
}
