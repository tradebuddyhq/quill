package stdlib

// GetUIRuntime returns a Tailwind-inspired utility CSS framework for Quill.
// Injected when components use className props or when explicitly imported.
func GetUIRuntime() string {
	return "// Quill UI , Utility CSS Framework\n" +
		"(function() {\n" +
		"  var css = " + quillUICSS + ";\n" +
		"  var style = document.createElement('style');\n" +
		"  style.id = 'quill-ui';\n" +
		"  style.textContent = css;\n" +
		"  if (document.head) document.head.appendChild(style);\n" +
		"  else document.addEventListener('DOMContentLoaded', function() { document.head.appendChild(style); });\n" +
		"})();\n"
}

// quillUICSS is the utility CSS as a JSON-encoded string constant for embedding in JS.
var quillUICSS = `"` +
	`/* Quill UI , Utility CSS Framework */` +
	`*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }` +
	`body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.5; color: #1a1a2e; }` +

	// Display
	`.block { display: block; }` +
	`.inline { display: inline; }` +
	`.inline-block { display: inline-block; }` +
	`.flex { display: flex; }` +
	`.inline-flex { display: inline-flex; }` +
	`.grid { display: grid; }` +
	`.hidden { display: none; }` +

	// Flex
	`.flex-row { flex-direction: row; }` +
	`.flex-col { flex-direction: column; }` +
	`.flex-wrap { flex-wrap: wrap; }` +
	`.flex-nowrap { flex-wrap: nowrap; }` +
	`.flex-1 { flex: 1 1 0%; }` +
	`.flex-auto { flex: 1 1 auto; }` +
	`.flex-none { flex: none; }` +
	`.grow { flex-grow: 1; }` +
	`.shrink-0 { flex-shrink: 0; }` +

	// Alignment
	`.items-start { align-items: flex-start; }` +
	`.items-center { align-items: center; }` +
	`.items-end { align-items: flex-end; }` +
	`.items-stretch { align-items: stretch; }` +
	`.justify-start { justify-content: flex-start; }` +
	`.justify-center { justify-content: center; }` +
	`.justify-end { justify-content: flex-end; }` +
	`.justify-between { justify-content: space-between; }` +
	`.justify-around { justify-content: space-around; }` +
	`.justify-evenly { justify-content: space-evenly; }` +
	`.self-start { align-self: flex-start; }` +
	`.self-center { align-self: center; }` +
	`.self-end { align-self: flex-end; }` +

	// Grid
	`.grid-cols-1 { grid-template-columns: repeat(1, minmax(0, 1fr)); }` +
	`.grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }` +
	`.grid-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); }` +
	`.grid-cols-4 { grid-template-columns: repeat(4, minmax(0, 1fr)); }` +
	`.grid-cols-6 { grid-template-columns: repeat(6, minmax(0, 1fr)); }` +
	`.grid-cols-12 { grid-template-columns: repeat(12, minmax(0, 1fr)); }` +
	`.col-span-2 { grid-column: span 2 / span 2; }` +
	`.col-span-3 { grid-column: span 3 / span 3; }` +
	`.col-span-4 { grid-column: span 4 / span 4; }` +
	`.col-span-6 { grid-column: span 6 / span 6; }` +

	// Gap
	`.gap-0 { gap: 0; }` +
	`.gap-1 { gap: 0.25rem; }` +
	`.gap-2 { gap: 0.5rem; }` +
	`.gap-3 { gap: 0.75rem; }` +
	`.gap-4 { gap: 1rem; }` +
	`.gap-5 { gap: 1.25rem; }` +
	`.gap-6 { gap: 1.5rem; }` +
	`.gap-8 { gap: 2rem; }` +
	`.gap-10 { gap: 2.5rem; }` +
	`.gap-12 { gap: 3rem; }` +

	// Padding
	`.p-0 { padding: 0; }` +
	`.p-1 { padding: 0.25rem; }` +
	`.p-2 { padding: 0.5rem; }` +
	`.p-3 { padding: 0.75rem; }` +
	`.p-4 { padding: 1rem; }` +
	`.p-5 { padding: 1.25rem; }` +
	`.p-6 { padding: 1.5rem; }` +
	`.p-8 { padding: 2rem; }` +
	`.p-10 { padding: 2.5rem; }` +
	`.p-12 { padding: 3rem; }` +
	`.px-0 { padding-left: 0; padding-right: 0; }` +
	`.px-1 { padding-left: 0.25rem; padding-right: 0.25rem; }` +
	`.px-2 { padding-left: 0.5rem; padding-right: 0.5rem; }` +
	`.px-3 { padding-left: 0.75rem; padding-right: 0.75rem; }` +
	`.px-4 { padding-left: 1rem; padding-right: 1rem; }` +
	`.px-6 { padding-left: 1.5rem; padding-right: 1.5rem; }` +
	`.px-8 { padding-left: 2rem; padding-right: 2rem; }` +
	`.py-0 { padding-top: 0; padding-bottom: 0; }` +
	`.py-1 { padding-top: 0.25rem; padding-bottom: 0.25rem; }` +
	`.py-2 { padding-top: 0.5rem; padding-bottom: 0.5rem; }` +
	`.py-3 { padding-top: 0.75rem; padding-bottom: 0.75rem; }` +
	`.py-4 { padding-top: 1rem; padding-bottom: 1rem; }` +
	`.py-6 { padding-top: 1.5rem; padding-bottom: 1.5rem; }` +
	`.py-8 { padding-top: 2rem; padding-bottom: 2rem; }` +

	// Margin
	`.m-0 { margin: 0; }` +
	`.m-1 { margin: 0.25rem; }` +
	`.m-2 { margin: 0.5rem; }` +
	`.m-3 { margin: 0.75rem; }` +
	`.m-4 { margin: 1rem; }` +
	`.m-6 { margin: 1.5rem; }` +
	`.m-8 { margin: 2rem; }` +
	`.m-auto { margin: auto; }` +
	`.mx-auto { margin-left: auto; margin-right: auto; }` +
	`.my-2 { margin-top: 0.5rem; margin-bottom: 0.5rem; }` +
	`.my-4 { margin-top: 1rem; margin-bottom: 1rem; }` +
	`.my-6 { margin-top: 1.5rem; margin-bottom: 1.5rem; }` +
	`.my-8 { margin-top: 2rem; margin-bottom: 2rem; }` +
	`.mt-0 { margin-top: 0; }` +
	`.mt-1 { margin-top: 0.25rem; }` +
	`.mt-2 { margin-top: 0.5rem; }` +
	`.mt-4 { margin-top: 1rem; }` +
	`.mt-6 { margin-top: 1.5rem; }` +
	`.mt-8 { margin-top: 2rem; }` +
	`.mb-0 { margin-bottom: 0; }` +
	`.mb-1 { margin-bottom: 0.25rem; }` +
	`.mb-2 { margin-bottom: 0.5rem; }` +
	`.mb-4 { margin-bottom: 1rem; }` +
	`.mb-6 { margin-bottom: 1.5rem; }` +
	`.mb-8 { margin-bottom: 2rem; }` +
	`.ml-auto { margin-left: auto; }` +
	`.mr-auto { margin-right: auto; }` +

	// Width & Height
	`.w-full { width: 100%; }` +
	`.w-screen { width: 100vw; }` +
	`.w-auto { width: auto; }` +
	`.w-fit { width: fit-content; }` +
	`.min-w-0 { min-width: 0; }` +
	`.max-w-sm { max-width: 24rem; }` +
	`.max-w-md { max-width: 28rem; }` +
	`.max-w-lg { max-width: 32rem; }` +
	`.max-w-xl { max-width: 36rem; }` +
	`.max-w-2xl { max-width: 42rem; }` +
	`.max-w-3xl { max-width: 48rem; }` +
	`.max-w-4xl { max-width: 56rem; }` +
	`.max-w-5xl { max-width: 64rem; }` +
	`.max-w-6xl { max-width: 72rem; }` +
	`.max-w-7xl { max-width: 80rem; }` +
	`.max-w-full { max-width: 100%; }` +
	`.h-full { height: 100%; }` +
	`.h-screen { height: 100vh; }` +
	`.h-auto { height: auto; }` +
	`.min-h-screen { min-height: 100vh; }` +

	// Typography
	`.text-xs { font-size: 0.75rem; line-height: 1rem; }` +
	`.text-sm { font-size: 0.875rem; line-height: 1.25rem; }` +
	`.text-base { font-size: 1rem; line-height: 1.5rem; }` +
	`.text-lg { font-size: 1.125rem; line-height: 1.75rem; }` +
	`.text-xl { font-size: 1.25rem; line-height: 1.75rem; }` +
	`.text-2xl { font-size: 1.5rem; line-height: 2rem; }` +
	`.text-3xl { font-size: 1.875rem; line-height: 2.25rem; }` +
	`.text-4xl { font-size: 2.25rem; line-height: 2.5rem; }` +
	`.text-5xl { font-size: 3rem; line-height: 1; }` +
	`.text-6xl { font-size: 3.75rem; line-height: 1; }` +
	`.font-thin { font-weight: 100; }` +
	`.font-light { font-weight: 300; }` +
	`.font-normal { font-weight: 400; }` +
	`.font-medium { font-weight: 500; }` +
	`.font-semibold { font-weight: 600; }` +
	`.font-bold { font-weight: 700; }` +
	`.font-extrabold { font-weight: 800; }` +
	`.italic { font-style: italic; }` +
	`.uppercase { text-transform: uppercase; }` +
	`.lowercase { text-transform: lowercase; }` +
	`.text-left { text-align: left; }` +
	`.text-center { text-align: center; }` +
	`.text-right { text-align: right; }` +
	`.underline { text-decoration: underline; }` +
	`.line-through { text-decoration: line-through; }` +
	`.no-underline { text-decoration: none; }` +
	`.leading-none { line-height: 1; }` +
	`.leading-tight { line-height: 1.25; }` +
	`.leading-normal { line-height: 1.5; }` +
	`.leading-relaxed { line-height: 1.625; }` +
	`.truncate { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }` +

	// Colors
	`.text-white { color: #ffffff; }` +
	`.text-black { color: #000000; }` +
	`.text-gray-400 { color: #9ca3af; }` +
	`.text-gray-500 { color: #6b7280; }` +
	`.text-gray-600 { color: #4b5563; }` +
	`.text-gray-700 { color: #374151; }` +
	`.text-gray-800 { color: #1f2937; }` +
	`.text-gray-900 { color: #111827; }` +
	`.text-red-500 { color: #ef4444; }` +
	`.text-red-600 { color: #dc2626; }` +
	`.text-green-500 { color: #22c55e; }` +
	`.text-green-600 { color: #16a34a; }` +
	`.text-blue-500 { color: #3b82f6; }` +
	`.text-blue-600 { color: #2563eb; }` +
	`.text-indigo-500 { color: #6366f1; }` +
	`.text-indigo-600 { color: #4f46e5; }` +
	`.text-purple-500 { color: #a855f7; }` +
	`.text-pink-500 { color: #ec4899; }` +

	// Background Colors
	`.bg-transparent { background-color: transparent; }` +
	`.bg-white { background-color: #ffffff; }` +
	`.bg-black { background-color: #000000; }` +
	`.bg-gray-50 { background-color: #f9fafb; }` +
	`.bg-gray-100 { background-color: #f3f4f6; }` +
	`.bg-gray-200 { background-color: #e5e7eb; }` +
	`.bg-gray-300 { background-color: #d1d5db; }` +
	`.bg-gray-500 { background-color: #6b7280; }` +
	`.bg-gray-700 { background-color: #374151; }` +
	`.bg-gray-800 { background-color: #1f2937; }` +
	`.bg-gray-900 { background-color: #111827; }` +
	`.bg-red-50 { background-color: #fef2f2; }` +
	`.bg-red-100 { background-color: #fee2e2; }` +
	`.bg-red-500 { background-color: #ef4444; }` +
	`.bg-red-600 { background-color: #dc2626; }` +
	`.bg-green-50 { background-color: #f0fdf4; }` +
	`.bg-green-100 { background-color: #dcfce7; }` +
	`.bg-green-500 { background-color: #22c55e; }` +
	`.bg-green-600 { background-color: #16a34a; }` +
	`.bg-blue-50 { background-color: #eff6ff; }` +
	`.bg-blue-100 { background-color: #dbeafe; }` +
	`.bg-blue-500 { background-color: #3b82f6; }` +
	`.bg-blue-600 { background-color: #2563eb; }` +
	`.bg-blue-700 { background-color: #1d4ed8; }` +
	`.bg-indigo-50 { background-color: #eef2ff; }` +
	`.bg-indigo-500 { background-color: #6366f1; }` +
	`.bg-indigo-600 { background-color: #4f46e5; }` +
	`.bg-purple-50 { background-color: #faf5ff; }` +
	`.bg-purple-500 { background-color: #a855f7; }` +

	// Borders
	`.border { border: 1px solid #e5e7eb; }` +
	`.border-0 { border-width: 0; }` +
	`.border-2 { border-width: 2px; }` +
	`.border-t { border-top: 1px solid #e5e7eb; }` +
	`.border-b { border-bottom: 1px solid #e5e7eb; }` +
	`.border-gray-200 { border-color: #e5e7eb; }` +
	`.border-gray-300 { border-color: #d1d5db; }` +
	`.border-red-500 { border-color: #ef4444; }` +
	`.border-green-500 { border-color: #22c55e; }` +
	`.border-blue-500 { border-color: #3b82f6; }` +
	`.border-transparent { border-color: transparent; }` +

	// Border Radius
	`.rounded-none { border-radius: 0; }` +
	`.rounded-sm { border-radius: 0.125rem; }` +
	`.rounded { border-radius: 0.25rem; }` +
	`.rounded-md { border-radius: 0.375rem; }` +
	`.rounded-lg { border-radius: 0.5rem; }` +
	`.rounded-xl { border-radius: 0.75rem; }` +
	`.rounded-2xl { border-radius: 1rem; }` +
	`.rounded-full { border-radius: 9999px; }` +

	// Shadows
	`.shadow-sm { box-shadow: 0 1px 2px 0 rgba(0,0,0,0.05); }` +
	`.shadow { box-shadow: 0 1px 3px 0 rgba(0,0,0,0.1), 0 1px 2px -1px rgba(0,0,0,0.1); }` +
	`.shadow-md { box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1), 0 2px 4px -2px rgba(0,0,0,0.1); }` +
	`.shadow-lg { box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -4px rgba(0,0,0,0.1); }` +
	`.shadow-xl { box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1), 0 8px 10px -6px rgba(0,0,0,0.1); }` +
	`.shadow-none { box-shadow: none; }` +

	// Opacity
	`.opacity-0 { opacity: 0; }` +
	`.opacity-50 { opacity: 0.5; }` +
	`.opacity-75 { opacity: 0.75; }` +
	`.opacity-100 { opacity: 1; }` +

	// Overflow
	`.overflow-auto { overflow: auto; }` +
	`.overflow-hidden { overflow: hidden; }` +
	`.overflow-scroll { overflow: scroll; }` +
	`.overflow-x-auto { overflow-x: auto; }` +
	`.overflow-y-auto { overflow-y: auto; }` +

	// Position
	`.static { position: static; }` +
	`.fixed { position: fixed; }` +
	`.absolute { position: absolute; }` +
	`.relative { position: relative; }` +
	`.sticky { position: sticky; }` +
	`.inset-0 { top: 0; right: 0; bottom: 0; left: 0; }` +
	`.top-0 { top: 0; }` +
	`.right-0 { right: 0; }` +
	`.bottom-0 { bottom: 0; }` +
	`.left-0 { left: 0; }` +

	// Z-Index
	`.z-0 { z-index: 0; }` +
	`.z-10 { z-index: 10; }` +
	`.z-20 { z-index: 20; }` +
	`.z-30 { z-index: 30; }` +
	`.z-40 { z-index: 40; }` +
	`.z-50 { z-index: 50; }` +

	// Cursor
	`.cursor-pointer { cursor: pointer; }` +
	`.cursor-default { cursor: default; }` +
	`.cursor-not-allowed { cursor: not-allowed; }` +
	`.select-none { user-select: none; }` +

	// Transitions
	`.transition { transition: all 150ms cubic-bezier(0.4, 0, 0.2, 1); }` +
	`.transition-colors { transition: color, background-color, border-color 150ms cubic-bezier(0.4, 0, 0.2, 1); }` +
	`.transition-transform { transition: transform 150ms cubic-bezier(0.4, 0, 0.2, 1); }` +
	`.transition-opacity { transition: opacity 150ms cubic-bezier(0.4, 0, 0.2, 1); }` +
	`.duration-150 { transition-duration: 150ms; }` +
	`.duration-200 { transition-duration: 200ms; }` +
	`.duration-300 { transition-duration: 300ms; }` +

	// Hover states
	`.hover\\\\:bg-gray-100:hover { background-color: #f3f4f6; }` +
	`.hover\\\\:bg-gray-200:hover { background-color: #e5e7eb; }` +
	`.hover\\\\:bg-red-600:hover { background-color: #dc2626; }` +
	`.hover\\\\:bg-red-700:hover { background-color: #b91c1c; }` +
	`.hover\\\\:bg-green-600:hover { background-color: #16a34a; }` +
	`.hover\\\\:bg-green-700:hover { background-color: #15803d; }` +
	`.hover\\\\:bg-blue-600:hover { background-color: #2563eb; }` +
	`.hover\\\\:bg-blue-700:hover { background-color: #1d4ed8; }` +
	`.hover\\\\:bg-indigo-600:hover { background-color: #4f46e5; }` +
	`.hover\\\\:bg-indigo-700:hover { background-color: #4338ca; }` +
	`.hover\\\\:bg-purple-600:hover { background-color: #9333ea; }` +
	`.hover\\\\:text-blue-600:hover { color: #2563eb; }` +
	`.hover\\\\:text-gray-900:hover { color: #111827; }` +
	`.hover\\\\:underline:hover { text-decoration: underline; }` +
	`.hover\\\\:shadow-md:hover { box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1), 0 2px 4px -2px rgba(0,0,0,0.1); }` +
	`.hover\\\\:shadow-lg:hover { box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -4px rgba(0,0,0,0.1); }` +
	`.hover\\\\:scale-105:hover { transform: scale(1.05); }` +

	// Focus states
	`.focus\\\\:outline-none:focus { outline: none; }` +
	`.focus\\\\:ring-2:focus { box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.5); }` +
	`.focus\\\\:ring-blue-500:focus { box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.5); }` +
	`.focus\\\\:ring-indigo-500:focus { box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.5); }` +
	`.focus\\\\:border-blue-500:focus { border-color: #3b82f6; }` +

	// Pre-built component classes
	`.btn { display: inline-flex; align-items: center; justify-content: center; padding: 0.5rem 1rem; font-weight: 500; border-radius: 0.375rem; transition: all 150ms; cursor: pointer; border: none; font-size: 0.875rem; line-height: 1.25rem; }` +
	`.btn-primary { background-color: #3b82f6; color: white; }` +
	`.btn-primary:hover { background-color: #2563eb; }` +
	`.btn-secondary { background-color: #6b7280; color: white; }` +
	`.btn-secondary:hover { background-color: #4b5563; }` +
	`.btn-success { background-color: #22c55e; color: white; }` +
	`.btn-success:hover { background-color: #16a34a; }` +
	`.btn-danger { background-color: #ef4444; color: white; }` +
	`.btn-danger:hover { background-color: #dc2626; }` +
	`.btn-outline { background-color: transparent; border: 1px solid #d1d5db; color: #374151; }` +
	`.btn-outline:hover { background-color: #f3f4f6; }` +
	`.btn-ghost { background-color: transparent; color: #374151; }` +
	`.btn-ghost:hover { background-color: #f3f4f6; }` +
	`.btn-sm { padding: 0.25rem 0.75rem; font-size: 0.75rem; }` +
	`.btn-lg { padding: 0.75rem 1.5rem; font-size: 1rem; }` +

	`.input { display: block; width: 100%; padding: 0.5rem 0.75rem; border: 1px solid #d1d5db; border-radius: 0.375rem; font-size: 0.875rem; line-height: 1.25rem; transition: border-color 150ms, box-shadow 150ms; }` +
	`.input:focus { outline: none; border-color: #3b82f6; box-shadow: 0 0 0 2px rgba(59,130,246,0.2); }` +

	`.card { background-color: white; border-radius: 0.5rem; border: 1px solid #e5e7eb; box-shadow: 0 1px 3px 0 rgba(0,0,0,0.1); padding: 1.5rem; }` +
	`.card-header { padding-bottom: 1rem; border-bottom: 1px solid #e5e7eb; margin-bottom: 1rem; }` +
	`.card-footer { padding-top: 1rem; border-top: 1px solid #e5e7eb; margin-top: 1rem; }` +

	`.badge { display: inline-flex; align-items: center; padding: 0.125rem 0.625rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 500; }` +
	`.badge-blue { background-color: #dbeafe; color: #1d4ed8; }` +
	`.badge-green { background-color: #dcfce7; color: #15803d; }` +
	`.badge-red { background-color: #fee2e2; color: #b91c1c; }` +
	`.badge-yellow { background-color: #fef9c3; color: #a16207; }` +
	`.badge-gray { background-color: #f3f4f6; color: #374151; }` +

	`.alert { padding: 1rem; border-radius: 0.375rem; margin-bottom: 1rem; }` +
	`.alert-info { background-color: #dbeafe; color: #1e40af; border: 1px solid #93c5fd; }` +
	`.alert-success { background-color: #dcfce7; color: #166534; border: 1px solid #86efac; }` +
	`.alert-warning { background-color: #fef9c3; color: #854d0e; border: 1px solid #fde047; }` +
	`.alert-error { background-color: #fee2e2; color: #991b1b; border: 1px solid #fca5a5; }` +

	`.container { width: 100%; margin-left: auto; margin-right: auto; padding-left: 1rem; padding-right: 1rem; }` +
	`@media (min-width: 640px) { .container { max-width: 640px; } }` +
	`@media (min-width: 768px) { .container { max-width: 768px; } }` +
	`@media (min-width: 1024px) { .container { max-width: 1024px; } }` +
	`@media (min-width: 1280px) { .container { max-width: 1280px; } }` +

	// Responsive
	`@media (min-width: 640px) { .sm\\\\:flex { display: flex; } .sm\\\\:hidden { display: none; } .sm\\\\:block { display: block; } .sm\\\\:grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); } .sm\\\\:grid-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); } }` +
	`@media (min-width: 768px) { .md\\\\:flex { display: flex; } .md\\\\:hidden { display: none; } .md\\\\:block { display: block; } .md\\\\:grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); } .md\\\\:grid-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); } .md\\\\:grid-cols-4 { grid-template-columns: repeat(4, minmax(0, 1fr)); } }` +
	`@media (min-width: 1024px) { .lg\\\\:flex { display: flex; } .lg\\\\:hidden { display: none; } .lg\\\\:block { display: block; } .lg\\\\:grid-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); } .lg\\\\:grid-cols-4 { grid-template-columns: repeat(4, minmax(0, 1fr)); } }` +

	// Animations
	`@keyframes spin { to { transform: rotate(360deg); } }` +
	`@keyframes pulse { 50% { opacity: 0.5; } }` +
	`@keyframes bounce { 0%, 100% { transform: translateY(-25%); animation-timing-function: cubic-bezier(0.8,0,1,1); } 50% { transform: none; animation-timing-function: cubic-bezier(0,0,0.2,1); } }` +
	`@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }` +
	`@keyframes slideUp { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: none; } }` +
	`.animate-spin { animation: spin 1s linear infinite; }` +
	`.animate-pulse { animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite; }` +
	`.animate-bounce { animation: bounce 1s infinite; }` +
	`.animate-fade-in { animation: fadeIn 300ms ease-out; }` +
	`.animate-slide-up { animation: slideUp 300ms ease-out; }` +

	// List & Table
	`.list-none { list-style: none; }` +
	`.list-disc { list-style: disc; }` +
	`.border-collapse { border-collapse: collapse; }` +

	// SR only
	`.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; border: 0; }` +
	`"`
