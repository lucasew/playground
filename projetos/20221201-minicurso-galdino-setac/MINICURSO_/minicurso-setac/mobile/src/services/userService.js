import api from "../utils/api";
import headers from "../utils/headers";
import AsyncStorage from "@react-native-community/async-storage";
import { navigate } from "../utils/navigate";

export default function logar(user){
    return api
        .post("/session", user)
        .then(x => {
            AsyncStorage.setItem("@access_token", x.data.token)
            navigate("PedidoPage")
        })
        .catch(err => alert(err))
    
}