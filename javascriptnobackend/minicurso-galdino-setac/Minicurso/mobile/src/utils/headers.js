import AsyncStorage from "@react-native-community/async-storage";
import { navigate } from "./navigate";

export default async function config(){
    const token = await AsyncStorage.getItem("@access_token");
    // const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjVkYThjOTk3OWMxMThhMDcyZDYwODIzYiIsImlhdCI6MTU3MTM0Mjc0NCwiZXhwIjoxNTcxNDI5MTQ0fQ.gHXmZlHHAUwNMOOutps-yntd5VYz1InEaG2T_kyv100"

    if(!token)
        navigate("LoginPage")
    else
        return {
            Authorization : `Bearer ${token}`
        }
    return ''
}
