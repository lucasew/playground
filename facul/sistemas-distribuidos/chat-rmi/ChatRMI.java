import java.rmi.RemoteException;
import java.rmi.Remote;
import java.rmi.Naming;
import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;
import java.rmi.server.UnicastRemoteObject;

import java.util.Map;
import java.util.HashMap;
import java.util.ArrayList;

public class ChatRMI {
    public static void main (String[] args) {
        if (args.length == 0) {
            System.out.println("Falta parâmetro: server ou client?");
            return;
        }
        if (args[0].equals("server")) {
            Server.main(new String[0]); //TODO: pegar parâmetros se necessário
        } else if (args[0].equals("client")) {
            Client.main(new String[0]);
        }
    }
}

public interface IWhatsUT extends Remote {
    IWhatsUTSession login(String user, String password) throws RemoteException, WhatsUTInvalidCredentialsException;
    void signup(String user, String password) throws RemoteException, WhatsUTInvalidCredentialsException;
    void ping() throws RemoteException; // status page
}

public interface IChat extends Remote {
    void sendMessage(String mensagem) throws RemoteException;
    Message[] getMessages() throws RemoteException;
}

public interface IWhatsUTSession extends Remote {
    String whoami() throws RemoteException;
    IChat conversarCom(String usuario) throws RemoteException;
}


public class Client {
    public static void main (String[] args) {
        try {
            String registryURL = "rmi://localhost:42069/whatsut";
            IWhatsUT w = (IWhatsUT) Naming.lookup(registryURL);
            w.ping();
            IWhatsUTSession sess = w.login("lucasew", "hunter2");
            System.out.println(sess.whoami());
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}

class Server {
    public static void main (String[] args) {
        try {
            Registry reg = LocateRegistry.createRegistry(42069);
            IWhatsUT w = new WhatsUTImpl();
            reg.rebind("whatsut", w);
        } catch (RemoteException e) {
            e.printStackTrace();
            System.out.println("bruh");
        }
    }
}


class UTFChat {
    private long currentMessageId = 0;
    private long currentChatId = 0;
    private Map<String, Chat> salas = new Chat();
    private Map<String, Chat> usuarios = new Chat();
    private String getJoinKey(String a, String b) {
        if (a > b) {
            return String.format("%s,%s", a, b);
        } else {
            return String.format("%s,%s", b, a);
        }
    }
    Chat getPrivado(String a, String b) {
        String key = this.getJoinKey(a, b);
        if (!this.usuarios.containsKey(key)) {
            this.usuarios.put(key, new Chat());
        }
        return this.usuarios.get(key);
    }
    public UTFChat() {}
}

class Mensagem {
    public String usuario;
    public String mensagem;
    public Mensagem(String usuario, String mensagem) {
        this.usuario = usuario;
        this.mensagem = mensagem;
    }
}

class Chat {
    public ArrayList<Mensagem> mensagens = new ArrayList<>();
    void enviarMensagem(String usuario, String mensagem) {
        this.mensagens.add(new Mensagem(usuario, mensagem));
    }
    Mensagem[] receberMensagens() {
        return this.toArray()
    }
}

class WhatsUTChatWrapper extends UnicastRemoteObject implements IChat {
    private Chat chat;
    private String username;
    public WhatsUTChatWrapper(String username, Chat chat) throws RemoteException {
        this.chat = chat;
        this.username = username;
    }
    public void sendMessage(String message) throws RemoteException {
        this.chat.enviarMensagem(this.username, message);
    }
    public Message[] getMessages() throws RemoteException {
        return this.chat.receberMensagens();
    }
}


class WhatsUTInvalidCredentialsException extends Exception {
    public WhatsUTInvalidCredentialsException() {
        super("usuário ou senha inválido");
    }
}

class WhatsUTImpl extends UnicastRemoteObject implements IWhatsUT {
    public WhatsUTImpl() throws RemoteException {
        super();
    }
    public IWhatsUTSession login(String user, String password) throws RemoteException, WhatsUTInvalidCredentialsException {
        return new WhatsUTSessionImpl(user); // TODO: dont send this to production
    }
    public void signup(String user, String password) throws RemoteException, WhatsUTInvalidCredentialsException {}
    public void ping() throws RemoteException {}
    public IChat conversarCom(String usuario) throws RemoteException {
    }
}

class WhatsUTSessionImpl extends UnicastRemoteObject implements IWhatsUTSession {
    private String login;
    public WhatsUTSessionImpl(String login) throws RemoteException {
        super();
        this.login = login;
    }
    public String whoami() throws RemoteException {
        return this.login;
    }
}
