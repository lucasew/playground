/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat.server;

import biglucas.chat.rmi.IChat;
import biglucas.chat.rmi.IWhatsUTSession;
import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;

/**
 *
 * @author lucasew
 */
public class WhatsUTSession extends UnicastRemoteObject implements IWhatsUTSession {
    final String login;
    final Core core;
    public WhatsUTSession(String login, Core core) throws RemoteException {
        this.login = login;
        this.core = core;
    }
    @Override
    public String whoami() throws RemoteException {
        return this.login;
    }

    @Override
    public IChat chatWith(String usuario) throws RemoteException {
       return new WhatsUTChat(this.core.getPrivateConversation(this.login, usuario), this.login);
    }
    
    public String[] listUsers() throws RemoteException {
        return this.core.users();
    }
    
    public String[] listGroups() throws RemoteException {
        return this.core.getGroups();
    }

    @Override
    public IChat joinGroup(String key) throws RemoteException {
        return new WhatsUTChat(this.core.getGroup(key, this.login), this.login);
    }

    @Override
    public void ping() throws RemoteException {}
    
}
