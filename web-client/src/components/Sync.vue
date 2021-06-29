<template>
  <v-container>
    <v-row justify="center">
      <h1>{{ h1Message }}</h1>
    </v-row>
    <v-row justify="center">
      <v-date-picker 
        v-model="picker" 
        color="#04819E"
        @click:date="click"
        :disabled="progress"
      >
      </v-date-picker>
    </v-row>
    <v-row justify="center">
      <v-alert
        type="error"
        v-show="showError"
      >
      {{ errorMessage }}
      </v-alert>
      <v-progress-linear
        color="cyan lighten-5"
        indeterminate
        rounded
        height="5"
        v-show="progress"
        light
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
        progress: false,
        showError: false,
        errorMessage: '',
        h1Message: 'Pick a date',
    }),
    
    methods: {
      click: async function(date){
        this.progress = true
        this.showError = false
        this.errorMessage = ''
        this.h1Message = 'This might take a few seconds...'

        try {
          await this.axios.get(`http://localhost:4000/sync?date=${date}`, 
            {headers: {'Access-Control-Allow-Origin': `http://localhost:9999`}})
            
          this.$router.push({path:'/buyers'})

        } catch(err) {
          this.showError = true
          if (err.response) {
            this.errorMessage = `Server Error`          
          } else if (err.request) {
            this.errorMessage = `Network Error`          
          } else {
            this.errorMessage = `Client Error`
          }
        }
        this.h1Message = ''
        this.progress = false
      }
    }
  }
</script>
