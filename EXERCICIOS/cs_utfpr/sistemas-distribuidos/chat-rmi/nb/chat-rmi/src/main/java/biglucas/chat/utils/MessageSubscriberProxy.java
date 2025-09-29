/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package biglucas.chat.utils;

import biglucas.chat.rmi.IChatCallback;
import biglucas.chat.server.Message;
import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;

/**
 *
 * @author lucasew
 */
public class MessageSubscriberProxy extends UnicastRemoteObject implements IChatCallback {

    private final OnMessage om;
    public MessageSubscriberProxy(OnMessage om) throws RemoteException {
        this.om = om;
    }

    @Override
    public void handleMessage(Message message) throws RemoteException {
        this.om.onMessage(message);
    }
}
