/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package tk.biglucas.ex1;

import java.util.ArrayList;
import java.util.stream.Stream;

/**
 *
 * @author lucasew
 */
public class Database {
    ArrayList<Item> items;

    public Database() {
        items = new ArrayList<>();
    }
    
    public void add(Item item) {
        items.add(item);
    }
    
    private void print_items(Stream<Item> items) {
        items.forEach(item -> item.print());
    }
    
    public void list() {
        print_items(items.stream());
    }
    
    public void listCD() {
        print_items(items.stream().filter(item -> item.getClass() == CD.class));
    }
    
    public void listDVD() {
        print_items(items.stream().filter(item -> item.getClass() == DVD.class));
    }
}
