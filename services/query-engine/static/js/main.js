// ==========================================
// Boogle Search Engine - Main Bundle
// ==========================================
// This file bundles:
// - Alpine.js (reactive components)
// - HTMX (dynamic interactions)
// - Custom utilities
// ==========================================

// Import Alpine.js
import Alpine from 'alpinejs';
window.Alpine = Alpine;

// Import HTMX
import htmx from 'htmx.org';
window.htmx = htmx;

// Start Alpine after DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    Alpine.start();
});

// ==========================================
// Custom JavaScript utilities for Boogle
// ==========================================

// Custom HTMX event listeners
document.addEventListener('htmx:beforeRequest', function(event) {
    // Show loading indicator
    const target = event.detail.target;
    if (target) {
        target.classList.add('htmx-loading');
    }
});

document.addEventListener('htmx:afterRequest', function(event) {
    // Hide loading indicator
    const target = event.detail.target;
    if (target) {
        target.classList.remove('htmx-loading');
    }
});

// Keyboard shortcuts
document.addEventListener('keydown', function(event) {
    // Focus search input with '/' key
    if (event.key === '/' && !['INPUT', 'TEXTAREA'].includes(document.activeElement.tagName)) {
        event.preventDefault();
        const searchInput = document.querySelector('input[name="query"]');
        if (searchInput) {
            searchInput.focus();
        }
    }
});

// Smooth scroll for anchor links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        const href = this.getAttribute('href');
        if (href !== '#') {
            e.preventDefault();
            const target = document.querySelector(href);
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth'
                });
            }
        }
    });
});

// Add ripple effect to buttons
function createRipple(event) {
    const button = event.currentTarget;
    const ripple = document.createElement('span');
    const diameter = Math.max(button.clientWidth, button.clientHeight);
    const radius = diameter / 2;

    ripple.style.width = ripple.style.height = `${diameter}px`;
    ripple.style.left = `${event.clientX - button.offsetLeft - radius}px`;
    ripple.style.top = `${event.clientY - button.offsetTop - radius}px`;
    ripple.classList.add('ripple');

    const existingRipple = button.querySelector('.ripple');
    if (existingRipple) {
        existingRipple.remove();
    }

    button.appendChild(ripple);
}

// Apply ripple effect to all buttons with ripple class
document.querySelectorAll('.btn-ripple').forEach(button => {
    button.addEventListener('click', createRipple);
});

console.log('🔍 Boogle Search Engine initialized');

