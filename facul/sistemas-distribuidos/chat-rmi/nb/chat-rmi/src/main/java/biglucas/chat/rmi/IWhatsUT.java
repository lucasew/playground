/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat.rmi;

import biglucas.chat.rmi.exceptions.WhatsUTInvalidCredentialsException;
import java.rmi.Remote;
import java.rmi.RemoteException;

/**
 *
 * @author lucasew
 */
public interface IWhatsUT extends Remote {
    IWhatsUTSession login(String user, String password) throws RemoteException, WhatsUTInvalidCredentialsException;
    void signup(String user, String password) throws RemoteException, WhatsUTInvalidCredentialsException;
    void ping() throws RemoteException;
    
}
