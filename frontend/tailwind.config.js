/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          light: '#e1f5e6', // light green background like the dribbble shot
          DEFAULT: '#3da562', // primary green
          dark: '#2a7844', // hover state
        },
        surface: {
          DEFAULT: '#ffffff',
          muted: '#f9fafb',
        }
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
      }
    },
  },
  plugins: [],
}
