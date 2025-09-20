/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Interface.java to edit this template
 */
package biglucas.chat.rmi;

import biglucas.chat.rmi.exceptions.WhatsUTForbiddenActionException;
import biglucas.chat.server.Message;
import java.rmi.Remote;
import java.rmi.RemoteException;

/**
 *
 * @author lucasew
 */
public interface IChat extends Remote {
    void sendMessage(String message) throws RemoteException, WhatsUTForbiddenActionException;
    Message[] getMessages() throws RemoteException, WhatsUTForbiddenActionException;
    String[] getMembers() throws RemoteException, WhatsUTForbiddenActionException;
    String[] getPendentMembers() throws RemoteException, WhatsUTForbiddenActionException;
    void allowMember(String member) throws RemoteException, WhatsUTForbiddenActionException;
    void subscribe(IChatCallback callback) throws RemoteException;
}
