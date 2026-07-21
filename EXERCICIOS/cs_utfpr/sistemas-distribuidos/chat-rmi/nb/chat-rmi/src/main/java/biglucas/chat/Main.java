/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat;

import java.util.Arrays;

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

        String[] subArgs = Arrays.copyOfRange(args, 1, args.length);

        if (args[0].equals("server")) {
            Server.main(subArgs);
        } else if (args[0].equals("client")) {
            Client.main(subArgs);
        }

    }
}
