<template>
    <v-card 
        class="mx-auto"
        max-width="400"
        tile
        >
        <v-list-item 
            three-line
            v-for="item in items"
            :key="item.id"
            >
            <v-list-item-content>
                <v-list-item-title v-text="item.name"></v-list-item-title>
                <v-list-item-subtitle>Age: {{item.age}}</v-list-item-subtitle>
                <v-list-item-subtitle>ID: {{item.id}}</v-list-item-subtitle>
            </v-list-item-content>
        </v-list-item>
    </v-card>
</template>

<script>
export default {
    name: 'AllBuyers',

    data: () => ({
        items: []
    }),

    async created() {
        try{
            const response = await this.axios.get(`http://localhost:4000/buyers`, 
                {headers: {'Access-Control-Allow-Origin': `http://localhost:9999`}})
            
            this.items = response.data

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