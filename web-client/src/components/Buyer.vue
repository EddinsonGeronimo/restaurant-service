<template>
    <v-container>
    <h1>Transactions</h1>
    <v-card class="mx-auto" tile>
            <v-list-group
                :value="true"
                no-action
                sub-group
                v-for="item in transactions"
                :key="item.id"
                >
                <template v-slot:activator>
                    <v-list-item three-line>
                        <v-list-item-content>
                            <v-list-item-title>ID: {{item.id}}</v-list-item-title>
                            <v-list-item-subtitle>Device: {{item.device}}</v-list-item-subtitle>
                            <v-list-item-subtitle>IP: {{item.ip}}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                </template>
                <h3>Products</h3>
                <v-list-item
                    v-for="product in item.Products"
                    :key="product.name"
                    link
                    two-line
                    >
                    <v-list-item-content>
                        <v-list-item-title v-text="product.name"></v-list-item-title>
                        <v-list-item-subtitle>Price: {{product.price}}</v-list-item-subtitle>
                    </v-list-item-content>
                </v-list-item>
            </v-list-group>
    </v-card>
    <v-divider></v-divider>
    <h1>Buyers with same IP address</h1>
    <v-card class="mx-auto" tile>
        <v-list-item 
            two-line
            v-for="item in otherBuyers"
            :key="item.id"
            >
            <v-list-item-content>
                <v-list-item-title v-text="item.name"></v-list-item-title>
                <v-list-item-subtitle>Age: {{item.age}}</v-list-item-subtitle>
            </v-list-item-content>
        </v-list-item>
    </v-card>
    <v-divider></v-divider>
    <h1>Recommended products</h1>
    <v-card class="mx-auto" tile>
        <v-list-item
            v-for="item in recommendedProducts"
            :key="item.id"
            >
            <v-list-item-content>
                <v-list-item-title v-text="item.name"></v-list-item-title>
            </v-list-item-content>
        </v-list-item>
    </v-card>
    </v-container>
</template>

<script>
export default {
    name: 'Buyer',

    data: () => ({
        transactions: [],
        otherBuyers: [],
        recommendedProducts: []
    }),

    // must be removed
    async created() {
        try{
            const response = await this.axios.get(`http://localhost:4000/buyer?id=43689b82`, 
                {headers: {'Access-Control-Allow-Origin': `http://localhost:9999`}})
            
            this.transactions = response.data.buyertransactions[0]
            this.otherBuyers = response.data.hassameip
            this.recommendedProducts = response.data.Rproducts

        } catch(err) {
            if (err.response) {
                // client received an error response (5xx, 4xx)
                console.log("Server Error:", err)
            } else if (err.request) {
                // client never received a response, or request never left
                console.log("Network Error:", err)
            } else {
                console.log("Client Error:", err)
            }
        }
    }
}
</script>