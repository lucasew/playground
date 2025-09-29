/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat.server;

import biglucas.chat.rmi.IChat;
import biglucas.chat.rmi.IChatCallback;
import biglucas.chat.rmi.exceptions.WhatsUTForbiddenActionException;
import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
import java.util.ArrayList;

/**
 *
 * @author lucasew
 */
public class WhatsUTChat extends UnicastRemoteObject implements IChat {
    private Chat chat;
    private String login;
    private ArrayList<IChatCallback> subscribers = new ArrayList<>();
    
    public void broadcastSubscribers(Message msg) {
        for (IChatCallback subscriber : subscribers) {
            System.out.printf("broadcast %s %s", subscriber.toString(), msg.toString());
            try {
                subscriber.handleMessage(msg);
            } catch (Exception e) {
                this.subscribers.remove(subscriber);
                e.printStackTrace();
            }
        }
    }
    
    public WhatsUTChat (Chat chat, String login) throws RemoteException {
        this.chat = chat;
        this.login = login;
    } 
    @Override
    public void sendMessage(String message) throws RemoteException, WhatsUTForbiddenActionException {
        this.chat.sendMessage(this.login, message);
    }

    @Override
    public Message[] getMessages() throws RemoteException, WhatsUTForbiddenActionException {
        return this.chat.getMessages(this.login);
    }

    @Override
    public String[] getMembers() throws RemoteException, WhatsUTForbiddenActionException {
        return this.chat.getMembers(this.login);
    }

    @Override
    public String[] getPendentMembers() throws RemoteException, WhatsUTForbiddenActionException {
        return this.chat.getPendentMembers(this.login);
    }

    @Override
    public void allowMember(String member) throws RemoteException, WhatsUTForbiddenActionException {
        this.chat.allowMember(this.login, member);
    }

    @Override
    public void subscribe(IChatCallback callback) {
        this.subscribers.add(callback);
    }
    
}
