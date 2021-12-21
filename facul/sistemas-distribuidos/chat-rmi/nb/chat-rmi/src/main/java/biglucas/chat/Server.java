/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat;

import biglucas.chat.rmi.IWhatsUT;
import biglucas.chat.server.WhatsUT;
import java.rmi.RemoteException;
import java.rmi.registry.LocateRegistry;
import java.rmi.registry.Registry;

/**
 *
 * @author lucasew
 */
public class Server {
    public static void main(String[] args) {
        try {
            Registry reg = LocateRegistry.createRegistry(42069);
            IWhatsUT w = new WhatsUT();
            reg.rebind("whatsut", w);
        } catch (RemoteException e) {
            e.printStackTrace();
            System.out.println("bruh");
        }
    }
}
