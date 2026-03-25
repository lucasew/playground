package biglucas.utils;

public class ErrorReporter {
    public static void reportError(Throwable e) {
        // Centralized error reporting
        // E.g., send to Sentry if available.
        e.printStackTrace();
    }
}
