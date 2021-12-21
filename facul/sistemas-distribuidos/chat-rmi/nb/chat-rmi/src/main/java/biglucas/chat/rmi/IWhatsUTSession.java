/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Interface.java to edit this template
 */
package biglucas.chat.rmi;

import java.rmi.Remote;
import java.rmi.RemoteException;

/**
 *
 * @author lucasew
 */
public interface IWhatsUTSession extends Remote {
    String whoami() throws RemoteException;
    IChat chatWith(String usuario) throws RemoteException;
    IChat joinGroup(String key) throws RemoteException;
    void ping() throws RemoteException;
    String[] listUsers() throws RemoteException;
    String[] listGroups() throws RemoteException;
}
