/** @type {import('tailwindcss').Config} */
export default {
	content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}"],
	theme: {
		extend: {
			colors: {
				"craft-dark": "#0F2A3A",
				"craft-surface": "#183A4F",
				"craft-light": "#FFF7EF",
				"craft-text": "#FFF7EF",
				"craft-primary": "#FF4DA6",
				"craft-secondary": "#FFC93D",
				"craft-accent": "#FF6F5C",
			},
			boxShadow: {
				lowpoly: "4px 4px 0 #000000",
				"lowpoly-lg": "8px 8px 0 #000000",
				"lowpoly-sm": "2px 2px 0 #000000",
			},
			fontFamily: {
				game: ['"Press Start 2P"', "monospace"],
			},
			borderRadius: {
				sm: "2px",
			},
		},
	},
	safelist: ["btn-lowpoly", "panel-lowpoly", "input-lowpoly"],
	plugins: [],
};
