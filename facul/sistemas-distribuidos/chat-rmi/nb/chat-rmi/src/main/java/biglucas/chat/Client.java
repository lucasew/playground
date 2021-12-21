/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat;

import biglucas.chat.rmi.IWhatsUT;
import biglucas.chat.ui.Login;
import java.rmi.Naming;

/**
 *
 * @author lucasew
 */
public class Client {
    public static void main(String[] args) {
        try {
            String registryURL = "rmi://localhost:42069/whatsut";
            IWhatsUT server = (IWhatsUT) Naming.lookup(registryURL);
            server.ping();
            Login ui = new Login(server);
            ui.setVisible(true);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
