public class ErrorReporter {
    public static void reportError(Exception e) {
        System.err.println("An unexpected error occurred: " + e.getMessage());
        e.printStackTrace(System.err);
        // Integrate with Sentry or other centralized reporting tools here
    }
}
