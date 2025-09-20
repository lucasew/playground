/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Interface.java to edit this template
 */
package biglucas.chat.rmi;

import biglucas.chat.server.Message;
import java.rmi.Remote;
import java.rmi.RemoteException;

/**
 *
 * @author lucasew
 */
public interface IChatCallback extends Remote {
    void handleMessage(Message message) throws RemoteException;
}
