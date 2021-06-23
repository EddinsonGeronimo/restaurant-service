<template>
  <v-container>
    <v-row justify="center">
      <v-date-picker v-model="picker"></v-date-picker>
    </v-row>
  </v-container>
</template>

<script>
  export default {
    name: 'Sync',

    data () {
      return {
        picker: new Date().toISOString().substr(0, 10),
      }
    },
  }
</script>
