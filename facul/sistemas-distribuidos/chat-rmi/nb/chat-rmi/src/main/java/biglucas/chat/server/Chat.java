/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat.server;

import biglucas.chat.rmi.IChatCallback;
import biglucas.chat.rmi.exceptions.WhatsUTForbiddenActionException;
import biglucas.chat.utils.OnMessage;
import java.util.ArrayList;
import java.util.Arrays;

/**
 *
 * @author lucasew
 */
public class Chat {
    private final ArrayList<String> members = new ArrayList();
    private final ArrayList<String> pendentMembers = new ArrayList();
    private final ArrayList<Message> messages = new ArrayList();
    private final IChatCallback onMessage;
    public Chat(String ...members) {
        this.members.addAll(Arrays.asList(members));
        this.onMessage = null;
    }
    public Chat(IChatCallback onMessage, String ...members) {
        this.members.addAll(Arrays.asList(members));
        this.onMessage = onMessage;
    }
    private void mustAllowed(String user) throws WhatsUTForbiddenActionException {
        if (!this.members.contains(user)) {
            throw new WhatsUTForbiddenActionException();
        }
    }
    void sendMessage(String from, String message) throws WhatsUTForbiddenActionException {
        this.mustAllowed(from);
        Message msg = new Message(from, message);
        this.messages.add(msg);
        try {
        if (this.onMessage != null) this.onMessage.handleMessage(msg);
        System.out.printf("<%s> %s\n", from, message);
        } catch (Exception e) {}
        
        
    }
    Message[] getMessages(String from) throws WhatsUTForbiddenActionException {
        this.mustAllowed(from);
        return this.messages.toArray(Message[]::new);
    }
    String[] getMembers(String from) throws WhatsUTForbiddenActionException {
        this.mustAllowed(from);
        return this.members.toArray(String[]::new);
    }
    String[] getPendentMembers(String from) throws WhatsUTForbiddenActionException {
        this.mustAllowed(from);
        return this.pendentMembers.toArray(String[]::new);
    }
    void allowMember(String from, String member) throws WhatsUTForbiddenActionException {
        this.mustAllowed(member);
        this.pendentMembers.remove(member);
        this.members.add(member);
    }
}
