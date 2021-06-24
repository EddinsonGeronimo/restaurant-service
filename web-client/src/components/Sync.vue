<template>
  <v-container>
    <v-row justify="center">
      <h1>Pick a date</h1>
    </v-row>
    <v-row justify="center">
      <v-date-picker 
        v-model="picker" 
        color="#04819E"
        @click:date="click" 
      >
      </v-date-picker>
    </v-row>
    <v-row>
      <v-progress-linear
        color="cyan lighten-5"
        indeterminate
        rounded
        height="5"
        v-show="progress"
        >
      </v-progress-linear>      
    </v-row>
  </v-container>
</template>

<script>
  export default {
    name: 'Sync',

    data: () => ({
        picker: new Date().toISOString().substr(0, 10),
        progress: false
    }),
    
    methods: {
      click: async function(date){
        this.progress = true
        try {
          await this.axios.get(`http://localhost:4000/sync?date=${date}`, 
            {headers: {'Access-Control-Allow-Origin': `http://localhost:9999`}})
            
          alert(`Data sync successful.`)

        } catch(err) {
          if (err.response) {
            alert(`Server Error:${err}` )
          } else if (err.request) {
              alert(`Network Error:${err}`)
          } else {
              alert(`Client Error:${err}`)
          }
          
        }
        this.progress = false
      }
    }
  }
</script>
