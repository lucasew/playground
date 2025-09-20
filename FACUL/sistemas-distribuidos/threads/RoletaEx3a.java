public class RoletaEx3a {
    public static void main (String[] args) throws InterruptedException {
        Roleta central = new Roleta();
        Roleta r1 = new Roleta(central);
        Roleta r2 = new Roleta(central);
        new Thread(() -> {
            for (int i = 0; i < 40000000; i++) {
                r1.incr();
            }
        }).start();
        new Thread(() -> {
            for (int i = 0; i < 120000000; i++) {
                r2.incr();
            }
        }).start();
        while (true) {
            Thread.sleep(100);
            System.out.println(central.getContagem());
        }
    }
}

class Roleta {
    private int count;
    private Roleta central;
    public Roleta() {
        this.count = 0;
    }
    public Roleta(Roleta r) {
        this.count = 0;
        this.central = r;
    }

    public synchronized void incr() {
        if (central != null) {
            central.incr();
        }
        this.count++;
    }

    public synchronized int getContagem() {
        if (this.central != null) {
            return this.central.getContagem();
        }
        return this.count;
    }
}
