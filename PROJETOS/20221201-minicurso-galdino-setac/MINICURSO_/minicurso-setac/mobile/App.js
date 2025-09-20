import React from "react";

import Login from "./src/pages/Login";
import Pedido from "./src/pages/Pedido";
import Routes from "./src/routes";
import { setNavigator } from "./src/utils/navigate";

export default App = () => {
    return (<Routes ref={setNavigator}/>)
}