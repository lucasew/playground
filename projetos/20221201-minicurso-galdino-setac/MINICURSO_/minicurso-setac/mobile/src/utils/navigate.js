import { NavigationActions, StackActions } from "react-navigation";

let navigator;

export function setNavigator(ref) {
  navigator = ref;
}

export function navigate(routeName, params) {
  navigator.dispatch(
    NavigationActions.navigate({
      routeName,
      params
    })
  );
}

/** Reseta a rota Stack para a determinada rota*/
export function resetStackTo(routeName, params) {
  const resetAction = StackActions.reset({
    index: 0,
    actions: [NavigationActions.navigate({ routeName, params })]
  });

  navigator.dispatch(resetAction);
}