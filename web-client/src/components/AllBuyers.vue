<template>
    <v-container>
        <v-row justify="center">
            <v-progress-circular
                indeterminate
                color="primary"
                v-show="progress"
                >
            </v-progress-circular>
        </v-row>
        <v-row>
            <v-card 
                class="mx-auto"
                tile
                >
                <v-list-item 
                    three-line
                    v-for="item in items"
                    :key="item.id"
                    @click="click(item)"
                    >
                    <v-list-item-content>
                        <v-list-item-title v-text="item.name"></v-list-item-title>
                        <v-list-item-subtitle>Age: {{item.age}}</v-list-item-subtitle>
                    </v-list-item-content>
                </v-list-item>
            </v-card>
    </v-row>
    </v-container>
</template>

<script>
export default {
    name: 'AllBuyers',

    data: () => ({
        items: [],
        progress: false
    }),

    async mounted() {
        this.progress = true
        try{
            const response = await this.axios.get(`http://localhost:4000/buyers`, 
                {headers: {'Access-Control-Allow-Origin': `http://localhost:9999`}})
            
            this.items = response.data.q

        } catch(err) {
            if (err.response) {
                alert(`Server Error`)
            } else if (err.request) {
                alert(`Network Error`)
            } else {
                alert(`Client Error`)
            }
            this.$router.push({path:'/'})
        }
        this.progress = false
    },

    methods: {
        click: async function(item){
            this.$router.push({name:'BuyerView', params: {itemId: item.id}})
            //alert(`${item.id}`)
        }
    }
}
</script>

<style scoped>
.v-progress-circular {
    margin: 1rem;
}
</style>