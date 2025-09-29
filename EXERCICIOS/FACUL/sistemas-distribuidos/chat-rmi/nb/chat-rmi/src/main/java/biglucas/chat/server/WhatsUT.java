/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat.server;

import biglucas.chat.rmi.IWhatsUT;
import biglucas.chat.rmi.IWhatsUTSession;
import biglucas.chat.rmi.exceptions.WhatsUTInvalidCredentialsException;
import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;

/**
 *
 * @author lucasew
 */
public class WhatsUT extends UnicastRemoteObject implements IWhatsUT {
    private final Core core = new Core();
    public WhatsUT() throws RemoteException {}
    @Override
    public IWhatsUTSession login(String user, String password) throws RemoteException, WhatsUTInvalidCredentialsException {
        return new WhatsUTSession(core.login(user, password), core);
    }

    @Override
    public void signup(String user, String password) throws RemoteException, WhatsUTInvalidCredentialsException {
        core.signUp(user, password);        
    }

    @Override
    public void ping() throws RemoteException {}
    
}
