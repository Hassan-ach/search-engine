// ==========================================
// Boogle Search Engine - Main Bundle
// ==========================================
// This file bundles:
// - Alpine.js (reactive components)
// - HTMX (dynamic interactions)
// - Custom utilities
// ==========================================

// Import Alpine.js
import Alpine from "alpinejs";
window.Alpine = Alpine;

// Import HTMX
// import htmx from "htmx.org";
window.htmx = require("htmx.org");

// // Start Alpine after DOM is ready
document.addEventListener("DOMContentLoaded", () => {
    Alpine.start();
});

console.log("Boogle Search Engine initialized");
