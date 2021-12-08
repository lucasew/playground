import java.util.Random;
import java.util.Queue;
import java.util.ArrayList;
import java.util.LinkedList;
import java.util.Iterator;

interface FailableRunner {
    void run() throws Exception;
}

class Threadloop extends Thread {
    FailableRunner runner;
    public Threadloop(FailableRunner runner) {
        this.runner = runner;
    }
    public void run() {
        try {
            while(true) {
                this.runner.run();
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}

public class BarbeiroEx2 {
    public static Random r = new Random();
    private static String[] nomes = {
        "Ryder",
        "Woozie",
        "Pulaski",
        "Toreno",
        "CJ",
        "Kendl",
        "Catalina",
        "Emmet",
        "Maria",
        "Cesar",
        "OG Loc",
        "Tempenny",
        "The Truth",
        "Big Smoke",
        "João",
        "ZERO",
        "Claude Speed",
    };
    private static String getRandomName() {
        return BarbeiroEx2.nomes[BarbeiroEx2.r.nextInt(BarbeiroEx2.nomes.length)];
    }
    public static void main (String[] args) throws InterruptedException {
        Iterator<Long> interval = BarbeiroEx2.r.longs(100, 800).iterator();
        Barbeiro b = new Barbeiro("Morgan Freeman");
        Barbearia br = new Barbearia();
        for (int i = 0; i < 4; i++) {
            br.construirNovaCadeira();
        }
        new Threadloop(() -> {
            br.barbeiroCortarCabelo(b);
        }).start();
        new Threadloop(() -> {
            Thread.sleep(interval.next());
            br.entrarNaFila(new Cliente(BarbeiroEx2.getRandomName()));
        }).start();
        new Threadloop(() -> {
            Thread.sleep(interval.next());
            br.entrarNaFila(new Cliente(BarbeiroEx2.getRandomName()));
        }).start();

    }
}

class Barbearia {
    private Queue<Cliente> fila;
    private ArrayList<Cadeira> cadeiras;
    public Barbearia() {
        this.fila = new LinkedList<>();
        this.cadeiras = new ArrayList<>();
    }

    private void tentarEscoarFila() {
        for(int i = 0; i < this.cadeiras.size(); i++) {
            Cadeira c = this.cadeiras.get(i);
            if (c.isOcupado) continue;
            if (this.fila.size() == 0) break;

            Cliente cl = fila.remove();
            System.out.printf("<cliente %s> * encontrou uma cadeira disponível *\n", cl.nome);
            c.isOcupado = true;
        }
        this.notify();
    }

    public synchronized void entrarNaFila(Cliente c) {
        System.out.printf("<cliente %s> * entrou na fila *\n", c.nome);
        fila.add(c);
        this.tentarEscoarFila();
    }

    public synchronized void barbeiroCortarCabelo(Barbeiro b) throws InterruptedException {
        boolean achou = false;
        for (int i = 0; i < this.cadeiras.size(); i++) {
            Cadeira c = this.cadeiras.get(i);
            if (c.isOcupado && !c.isCortando) {
                achou = true;
                System.out.printf("<barbeiro %s> está cortando cabelo\n", b.nome);
                c.isCortando = true;
                Thread.sleep(200);
                c.isOcupado = false;
                c.isCortando = false;
            }
        }
        if (!achou) {
            System.out.printf("<barbeiro %s> * a mimir *\n", b.nome);
            this.wait();
        }
        this.tentarEscoarFila();
    }
    public synchronized void construirNovaCadeira() {
        this.cadeiras.add(new Cadeira());
    }
}

class Cadeira {
    public boolean isOcupado = false;
    public boolean isCortando = false;
    public Cadeira() {}
}

class Barbeiro extends Pessoa {
    public Barbeiro(String nome) {
        super(nome);
    }
}

class Cliente extends Pessoa {
    public Cliente(String nome) {
        super(nome);
    }
}

class Pessoa {
    public String nome;
    public Pessoa(String nome) {
        this.nome = nome;
    }
}
