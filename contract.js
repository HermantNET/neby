const bot = "n1HY4ob2kBRacVHoJKGqmwsvhE6BN5pvEy2"

class Yo {
  constructor() {
    LocalContractStorage.defineMapProperty(this, "accounts")
  }

  init() {}

  getAccount(id) {
    if (Blockchain.transaction.from !== bot) throw new Error("unauthorized")
    const account = this.accounts.get(id)
    return account
  }

  setAccount(id, address) {
    if (Blockchain.transaction.from !== bot) throw new Error("unauthorized")
    this.accounts.set(id, address)
  }
}

module.exports = Yo
