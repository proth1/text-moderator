import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string | Date): string {
  return new Date(date).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatPercentage(value: number): string {
  return `${(value * 100).toFixed(1)}%`;
}

export function getCategoryColor(category: string): string {
  const colors: Record<string, string> = {
    hate_speech: "red",
    harassment: "orange",
    violence: "red",
    sexual_content: "purple",
    spam: "yellow",
    misinformation: "blue",
    pii: "green",
  };
  return colors[category] || "gray";
}
