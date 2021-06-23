<template>
  <v-container>
    <v-row justify="center">
      <v-date-picker 
        v-model="picker" 
        color="#04819E"
        @click:date="click" 
      >
      </v-date-picker>
    </v-row>
  </v-container>
</template>

<script>
  export default {
    name: 'Sync',

    data: () => ({
        picker: new Date().toISOString().substr(0, 10),
    }),
    methods: {
      click: async function(date, event){
        try {
          const response = await this.axios.get(`http://localhost:4000/sync`, 
            {headers: {'Access-Control-Allow-Origin': `http://localhost:9999`}})
          console.log(response.data.task + date + event)
          
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
  }
</script>
