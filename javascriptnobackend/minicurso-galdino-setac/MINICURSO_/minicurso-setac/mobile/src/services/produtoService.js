import api from "../utils/api";
import headers from '../utils/headers'
import Axios from "axios";

export async function buscarProdutos(){
    
    const config = await headers()
    return api.get('products',{headers: config});
}

export async function salvarComanda(command){
    const config = await headers()
    return api.post("orders", command, {headers: config});
}