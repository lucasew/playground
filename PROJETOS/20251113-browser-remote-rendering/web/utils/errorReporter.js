window.captureAppError = function(message, error) {
    // Centralized error reporting
    console.error('Error Reported:', message, error || '');
    // In a real app, send to Sentry or equivalent here
};
