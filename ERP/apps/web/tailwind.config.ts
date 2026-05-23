import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: "#6B3F1F",
          50:  "#FAF5F0",
          100: "#F0DCC9",
          500: "#6B3F1F",
          600: "#5A3418",
          700: "#472810",
        },
      },
    },
  },
  plugins: [],
} satisfies Config;
