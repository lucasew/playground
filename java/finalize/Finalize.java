public class Finalize {
    public static void main(String args[]) {
        System.out.println("inicio");
        Finalize res = new Finalize();
        res = null;
        System.gc();
        System.out.println("post gc");
    }
    @Override
    protected void finalize() throws Throwable {
        System.out.println("fim");
    }
}
