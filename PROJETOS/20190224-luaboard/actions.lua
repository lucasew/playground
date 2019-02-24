function dir(obj)
    for k, v in pairs(obj) do
        print(k)
    end
end

while true do
    print("Mana: ", me.get_mana())
    print("GoAhead: ", action.go_ahead(2))
    print("GetPos: ", me.get_pos())
    rnd = rand(200)
    print("RANDOM: ", rnd)
    print("Round: ", round(rnd))
    delay(1000)
end
