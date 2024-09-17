if __name__ == "__main__":
    f = open("log.txt", "r")

    unknown_count = 0
    light_count = 0
    bridge_count = 0
    full_count = 0
    celestia_celestia_count = 0
    nebula_count = 0
    lumina_count = 0
    libp2p_count = 0
    for line in f.readlines():
        agent = line.split(" ")[-1][:-1]

        if agent == "unknown":
            unknown_count += 1
        elif "celestia-node/celestia/light" in agent:
            light_count += 1
        elif "celestia-node/celestia/bridge" in agent:
            bridge_count += 1
        elif "celestia-node/celestia/full" in agent:
            full_count += 1
        elif "celestia-celestia" in agent:
            celestia_celestia_count += 1
        elif "nebula" in agent:
            nebula_count += 1
        elif "lumina/celestia" in agent:
            lumina_count += 1
        elif "rust-libp2p" in agent:
            libp2p_count += 1
        elif agent != "":
            print(agent)

    print("unknown:", unknown_count)
    print("light:", light_count)
    print("bridge:", bridge_count)
    print("full:", full_count)
    print("celestia-celestia:", celestia_celestia_count)
    # print("nebula:", nebula_count)
    print("lumina:", lumina_count)
    print("rust-libp2p:", libp2p_count)
    print(
        "total:",
        unknown_count
        + light_count
        + bridge_count
        + full_count
        + celestia_celestia_count
        # + nebula_count
        + lumina_count
        + libp2p_count,
    )
