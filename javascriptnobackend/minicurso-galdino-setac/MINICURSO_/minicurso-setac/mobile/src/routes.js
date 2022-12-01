import { createAppContainer, createSwitchNavigator } from "react-navigation";
import { createStackNavigator, StackViewTransitionConfigs } from "react-navigation-stack";
import Login from "./pages/Login";
import Pedido from "./pages/Pedido";


const routes = createStackNavigator (
    {
      LoginPage: {
        screen: Login
      },
      PedidoPage: {
        screen: Pedido
      }
    },
    {
      headerMode: "none",
      initialRouteName: "LoginPage"
    }
  )
 
export default createAppContainer(routes);