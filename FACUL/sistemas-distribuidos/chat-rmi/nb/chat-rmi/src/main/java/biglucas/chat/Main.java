/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat;

/**
 *
 * @author lucasew
 */
public class Main {

    public static void main(String[] args) {
        if (args.length == 0) {
            System.out.println("Falta parâmetro: server ou client?");
            return;
        }
        if (args[0].equals("server")) {
            Server.main(new String[0]); // TODO: fatiar parametros
        } else if (args[0].equals("client")) {
            Client.main(new String[0]); // TODO: fatiar parâmetros
        }

    }
}
