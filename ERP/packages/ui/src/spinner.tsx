import * as React from "react";
import { cn } from "./utils";

interface SpinnerProps {
  size?: "sm" | "md" | "lg";
  className?: string;
}

const sizes = { sm: "h-4 w-4", md: "h-6 w-6", lg: "h-10 w-10" };

export function Spinner({ size = "md", className }: SpinnerProps) {
  return (
    <div
      role="status"
      className={cn(
        "animate-spin rounded-full border-2 border-gray-300 border-t-brand",
        sizes[size],
        className
      )}
      aria-label="Loading"
    />
  );
}
