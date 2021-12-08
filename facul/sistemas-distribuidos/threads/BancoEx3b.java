import java.util.Map;
import java.util.HashMap;

public class BancoEx3b {
    public static void main (String[] args) {
        Banco bank = new Banco();
        InternetBanking a = new InternetBanking(bank, "alice");
        InternetBanking b = new InternetBanking(bank, "bob");
        a.deposito(30d);
        b.deposito(30d);
        a.transferencia("bob", 20d);
        System.out.printf("alice: %f\n", a.getSaldo());
        System.out.printf("bob: %f\n", b.getSaldo());
    }
}

class Banco {
    private Map<String, Double> saldos = new HashMap<>();
    public Banco() {}

    public synchronized void deposito(String conta, Double valor) {
        if (valor <= 0) return;
        Double saldoAntigo = this.getSaldo(conta);
        this.saldos.put(conta, saldoAntigo + valor);
    }
    public synchronized void saque(String conta, Double valor) {
        if (valor <= 0) return;
        Double saldoAntigo = this.getSaldo(conta);
        if (saldoAntigo - valor < 0) return;
        this.saldos.put(conta, saldoAntigo - valor);
    }

    public synchronized Double getSaldo(String conta) {
        return this.saldos.getOrDefault(conta, 0d);
    }
    
    public synchronized void transferencia(String origem, String destino, Double valor) {
        if (valor <= 0) return;
        if (this.getSaldo(origem) < valor) return;
        Double origemSaldo = this.getSaldo(origem);
        Double destinoSaldo = this.getSaldo(destino);
        this.saldos.put(origem, origemSaldo - valor);
        this.saldos.put(destino, destinoSaldo + valor);
    }
}

class InternetBanking {
    private Banco b;
    private String conta;
    public InternetBanking(Banco b, String conta) {
        this.b = b;
        this.conta = conta;
    }

    public void deposito(Double valor) {
        this.b.deposito(this.conta, valor);
    }

    public void saque(Double valor) {
        this.b.saque(this.conta, valor);
    }

    public Double getSaldo() {
        return this.b.getSaldo(this.conta);
    }

    public void transferencia(String destino, Double valor) {
        this.b.transferencia(this.conta, destino, valor);
    }
}
