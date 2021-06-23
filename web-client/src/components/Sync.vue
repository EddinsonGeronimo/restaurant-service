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
    methods:{
      async click(date){
        alert(`date clicked: ${date}`)
      }
    }
  }
</script>
