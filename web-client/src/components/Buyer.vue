<template>
    <v-container>
    <v-row
        justify="center"
        >
        <v-col
            cols="1"
            >
            <v-btn
                :loading="loading"
                :disabled="loading"
                color="blue-grey"
                class="ma-2 white--text"
                fab
                @click="click"
            >
                <v-icon dark>
                    mdi-cloud-download
                </v-icon>
            </v-btn> 
        </v-col>
        <v-col
            cols="2"
            >
            <v-text-field
                label="Buyer id"
                v-model="buyerid"
                >
            </v-text-field>
        </v-col>
    </v-row>
    <h1 v-show="progress">Buyer</h1>
    <h3 v-show="progress">Name: {{buyername}}</h3>
    <h3 v-show="progress">Age: {{buyerage}}</h3>
    <v-divider v-show="progress"></v-divider>
    <h1 v-show="progress">Transactions</h1>
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
                    v-for="product in item.products"
                    :key="product.name"
                    link
                    two-line
                    >
                    <v-list-item-content>
                        <v-list-item-title>Name: {{product.name}}</v-list-item-title>
                        <v-list-item-subtitle>Price: {{product.price}}</v-list-item-subtitle>
                    </v-list-item-content>
                </v-list-item>
            </v-list-group>
    </v-card>
    <v-divider v-show="progress"></v-divider>
    <h1 v-show="progress">Buyers with same IP address</h1>
    <v-card class="mx-auto" tile>
        <v-list-item 
            v-for="item in otherBuyers"
            :key="item.id"
            >
            <v-list-item-content>
                <v-list-item-title v-text="item.name"></v-list-item-title>
            </v-list-item-content>
        </v-list-item>
    </v-card>
    <v-divider v-show="progress"></v-divider>
    <h1 v-show="progress">Recommended products</h1>
    <v-card class="mx-auto" tile>
        <v-list-item
            v-for="item in recommendedProducts"
            :key="item.id"
            two-line
            >
            <v-list-item-content>
                <v-list-item-title>Name: {{item.name}}</v-list-item-title>
                <v-list-item-subtitle>ID: {{item.id}}</v-list-item-subtitle>
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
        recommendedProducts: [],
        loading: false,
        buyerid: '',
        buyername: '',
        buyerage: '',
        progress: false
    }),

    async created(){
        this.loadBuyer()
    },

    watch: {
        '$route': 'loadBuyer'
    },

    methods: {
        click: async function(){
            
            if(this.buyerid.length == 0) {
                alert(`The 'Buyer id' field is required.`)
                return
            }
            this.loading = true
            try{
                const response = await this.axios.get(`http://localhost:4000/buyers/${this.buyerid}`, 
                {headers: {'Access-Control-Allow-Origin': `http://localhost:9999`}})
                
                this.transactions = response.data.buyerandtrans[0].transactions
                this.otherBuyers = response.data.hassameip
                this.recommendedProducts = response.data.rproducts

                this.progress = true
            }
            catch(err){
                if (err.response) {
                    alert(`Server Error`)
                } else if (err.request) {
                    alert(`Network Error`)
                } else {
                    alert(`Client Error`)
                }
                this.$router.push({path:'/'})
            }
            this.loading = false
        },

        loadBuyer: async function(){
            this.buyerid = this.$route.params.itemId
            this.buyername = this.$route.params.itemName
            this.buyerage = this.$route.params.itemAge
            this.click()
        }
    }
}
</script>