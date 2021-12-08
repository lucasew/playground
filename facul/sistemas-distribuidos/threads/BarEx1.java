import java.util.HashMap;
import java.util.Random;
import java.util.Iterator;

public class BarEx1 {
    public static void main (String[] args) throws InterruptedException {
        int numClientes = 100;
        int numGarcons = 3;
        int numRodadas = 3;
        Bar bar = new Bar(numRodadas);
        for (int i = 0; i < numClientes; i++) {
            new Thread(new Consumidor(bar)).start();
        }
        Thread.sleep(3000);
        for (int i = 0; i < numGarcons; i++) {
            new Thread(new Garcom(bar)).start();
        }
    }
}

class Ingresso {
    int id;
    public Ingresso(int id) {
        this.id = id;
    }
}

class Garcom implements Runnable {
    Bar bar;
    Iterator<Long> random = new Random().longs(200, 500).iterator(); // waiter wait time
    Garcom(Bar bar) {
        this.bar = bar;
    }
    public void run() {
        while(true) {
            try {
                Thread.sleep(random.next());
                bar.garcom_entregar_pedido();
                System.out.println("<garçom> entregue pedido");
            } catch (InterruptedException e) {
                System.out.println("<garçom> * a mimir *");
            } catch (AcabouOGoleException e) {
                System.out.println("<garçom> trabalho finalizado!");
                break;
            }
        }
    }
}

class Consumidor implements Runnable {
    private Bar bar;
    private Ingresso ingresso; Iterator<Long> random = new Random().longs(1000, 2500).iterator(); // waiter wait time
    public Consumidor(Bar bar) {
        this.bar = bar;
        this.ingresso = bar.comprar_ingresso();
    }

    public void run() {
        while(true) {
            try {
                Thread.sleep(random.next());
                bar.pedir_gole(this.ingresso);
            } catch (InterruptedException e) {
                System.out.printf("<consumidor %d> * capotou o corsa *\n", this.ingresso.id);
            } catch (AcabouOGoleException e) {
                System.out.printf("<consumidor %d> nem queria mesmo kkkkj\n", this.ingresso.id);
                break;
            }
        }
    }
}

class Bar {
    private int cliente_id = 0;
    private HashMap<Integer, Boolean> pediu = new HashMap<Integer, Boolean>();
    private int pedidos_restantes;
    private int rodadas_restantes;

    public Bar(int rodadas) {
        this.rodadas_restantes = rodadas;
    }

    public synchronized void pedir_gole(Ingresso i) throws AcabouOGoleException {
        if (this.rodadas_restantes == 0) {
            throw new AcabouOGoleException();
        }
        if (this.pediu.containsValue(i.id)) {
            return;
        }
        System.out.printf("<cliente %d> manda mais ai o/\n", i.id);
        this.pediu.put(i.id, true);
    }
    public synchronized Ingresso comprar_ingresso() {
        this.cliente_id++;
        return new Ingresso(this.cliente_id);
    }
    public synchronized void garcom_entregar_pedido() throws AcabouOGoleException {
        if (this.pedidos_restantes <= 0) {
            if (this.rodadas_restantes == 0) {
                throw new AcabouOGoleException();
            }
            System.out.println("<system> próxima rodada liberada");
            this.pedidos_restantes = this.pediu.size();
            this.rodadas_restantes--;
            this.pediu.clear();
        }
        System.out.printf("<system> pedidos restantes: %d\n", this.pedidos_restantes);
        this.pedidos_restantes--;
    }
}

class AcabouOGoleException extends Exception {
    public AcabouOGoleException() {
        super("Acabou o gole");
    }
}
