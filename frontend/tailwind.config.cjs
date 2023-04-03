/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      animation: {
        'wiggle': 'wiggle 0.5s linear infinite',
      },
      keyframes: {
        wiggle: {
          '0%, 80%, 98%, 99%, 100%': { transform: 'rotate(0deg)' },
          '25%': { transform: 'rotate(8deg)' },
          '75%': { transform: 'rotate(-8deg)' },
        }
      },
    },
  },
}
