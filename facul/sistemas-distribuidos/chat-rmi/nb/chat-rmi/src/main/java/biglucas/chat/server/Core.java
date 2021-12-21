/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat.server;

import biglucas.chat.rmi.exceptions.WhatsUTInvalidCredentialsException;
import java.util.Map;
import java.util.HashMap;
import java.util.stream.StreamSupport;
import java.util.Spliterators;

/**
 *
 * @author lucasew
 */
public class Core {
    private Map<String, String> users = new HashMap<>();
    private Map<String, Chat> chats = new HashMap<>();
    private String getPrivateChatKey(String a, String b) {
        if (a.compareTo(b) > 0) {
            return String.format("%s;%s", a, b);
        } else {
            return String.format("%s;%s", a, b);
        }
    }
    private Chat getOrCreateChat(String key, String ...users) {
        if (!this.chats.containsKey(key)) {
            Chat c = new Chat(users);
            this.chats.put(key, c);
        }
        return this.chats.get(key);
    }
    public Chat getPrivateConversation(String user, String anotherUser) {
        String key = "p" + this.getPrivateChatKey(user, anotherUser);
        return this.getOrCreateChat(key, user, anotherUser);
    }
    public Chat getGroup(String key, String ...user) {
        return this.getOrCreateChat("g" + key, user);
    }
    public String login(String user, String password) throws WhatsUTInvalidCredentialsException {
        if (!this.users.containsKey(user)) {
            throw new WhatsUTInvalidCredentialsException();
        }
        String pass = this.users.get(user);
        if (pass.equals(password)) {
            return user;
        }
        throw new WhatsUTInvalidCredentialsException();
    }
    public String signUp(String user, String password) throws WhatsUTInvalidCredentialsException {
        if (this.users.containsKey(user)) {
            throw new WhatsUTInvalidCredentialsException();
        }
        this.users.put(user, password);
        return user;
    }
    
    public String[] users() {
        return StreamSupport.stream(
                Spliterators.spliteratorUnknownSize(
                        this.users.keySet().iterator(), 0
                ), false)
            .sorted().toArray(String[]::new);
    }

    String[] getGroups() {
        return StreamSupport.stream(
                Spliterators.spliteratorUnknownSize(
                        this.chats.keySet().iterator(), 0
                ), false)
            .sorted()
            .filter((s) -> s.charAt(0) == 'g')
            .map((s) -> s.substring(1))
            .toArray(String[]::new);
    }
}
