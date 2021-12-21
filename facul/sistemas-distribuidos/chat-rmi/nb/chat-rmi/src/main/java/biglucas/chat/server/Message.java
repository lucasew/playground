/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat.server;

import java.io.Serializable;

/**
 *
 * @author lucasew
 */
public class Message implements Serializable {
    public String text;
    public String sender;
    public Message(String sender, String text) {
        this.text = text;
        this.sender = sender;
    }
}
